package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
)

// Executor runs a Query against an Input Table
type Executor struct {
	Pretty bool
}

func NewExecutor() *Executor {
	return &Executor{
		Pretty: false,
	}
}

func (e *Executor) Execute(q *Query, input database.Table, w io.Writer) error {
	// Build the finalized table plan (resolving FROM subqueries and applying WHERE/GROUP/SELECT)
	finalTable, err := e.BuildTable(q, input)
	if err != nil {
		return err
	}

	// Iterate and Print Results
	iterator, err := finalTable.Iterate()
	if err != nil {
		return err
	}
	defer iterator.Close()

	// Stream results as JSONL
	encoder := json.NewEncoder(w)
	if e.Pretty {
		encoder.SetIndent("", "  ")
	} else {
		encoder.SetIndent("", "")
	}

	for iterator.Next() {
		row := iterator.Row().Primitive()
		if err := encoder.Encode(row); err != nil {
			return err
		}
	}

	if err := iterator.Error(); err != nil {
		return err
	}

	return nil
}

// BuildTable constructs the logical table for a query, handling recursion for subqueries
func (e *Executor) BuildTable(q *Query, input database.Table) (database.Table, error) {
	// 1. Resolve Input Source (FROM clause or default input)
	var currentTable database.Table = input

	if q.From != "" {
		// Check if it's a subquery (starts with SELECT)
		if strings.HasPrefix(strings.ToUpper(q.From), "SELECT") {
			// Recursive Parse & Build
			subQ, err := ParseQuery(q.From)
			if err != nil {
				return nil, fmt.Errorf("failed to parse subquery: %w", err)
			}
			// Execute subquery using the SAME input (inheritance)
			// Or should it be fresh? Usually subqueries in FROM don't inherit row context
			// unless lateral. But here we assume it's just a derived table on the same source data keys?
			// Actually, standard SQL: FROM (SELECT ... FROM table)
			// If inner has no FROM, what does it use?
			// In our CLI context, "FROM" usually implies overriding the source.
			// But if the inner query has NO FROM, it should default to the file input.
			// So passing `input` down is correct.
			subTable, err := e.BuildTable(subQ, input)
			if err != nil {
				return nil, err
			}
			currentTable = subTable
		} else {
			// TODO: Support file path in FROM?
			// For now, treat as error or ignore
			// fmt.Printf("Warning: FROM '%s' not supported (only subqueries)\n", q.From)
		}
	}

	// 2. Refine query to handle $ operator
	if q.Condition != "" {
		refineQuery(q)
	}

	// 3. Apply WHERE (Filter)
	if q.Condition != "" {
		expr := query.ParseExpression(q.Condition)
		currentTable = &FilterTable{
			source:     currentTable,
			expression: expr,
		}
	}

	// 4. Apply SELECT (Projection) or Aggregation
	hasAggregation := q.GroupBy != ""
	if !hasAggregation {
		for _, f := range q.Fields {
			if f.Aggregate != "" {
				hasAggregation = true
				break
			}
		}
	}

	if hasAggregation {
		currentTable = &AggregateTable{
			source: currentTable,
			query:  q,
		}
	} else if len(q.Fields) > 0 {
		currentTable = &ProjectTable{
			source: currentTable,
			fields: q.Fields,
		}
	}

	return currentTable, nil
}

// FilterTable wraps a source table and filters rows
type FilterTable struct {
	source     database.Table
	expression query.Expression
}

func (t *FilterTable) Iterate() (database.RowIterator, error) {
	srcIter, err := t.source.Iterate()
	if err != nil {
		return nil, err
	}
	return &filterIterator{source: srcIter, expression: t.expression}, nil
}

type filterIterator struct {
	source     database.RowIterator
	expression query.Expression
}

func (it *filterIterator) Next() bool {
	for it.source.Next() {
		// Convert Row back to Record for Match
		// This is a bit inefficient (interface roundtrip), optimization for later
		primitive := it.source.Row().Primitive()

		// Attempt to convert primitive to Record-compatible map or handle directly
		// query.Filter.Match expects parser.Record (map[string]interface{})
		var record map[string]interface{}
		switch v := primitive.(type) {
		case parser.Record:
			record = v
		case map[string]interface{}:
			record = v
		case database.OrderedMap:
			record = v.ToMap()
		default:
			// If not a map, wrap it? Or skip?
			// The filter logic might need adjustment for primitives but currently assumes maps.
			continue
		}

		if it.expression.Evaluate(record) {
			return true
		}
	}
	return false
}

func (it *filterIterator) Row() database.Row {
	return it.source.Row()
}

func (it *filterIterator) Error() error {
	return it.source.Error()
}

func (it *filterIterator) Close() error {
	return it.source.Close()
}

// ProjectTable wraps a source table and selects specific fields
type ProjectTable struct {
	source database.Table
	fields []Field
}

func (t *ProjectTable) Iterate() (database.RowIterator, error) {
	srcIter, err := t.source.Iterate()
	if err != nil {
		return nil, err
	}
	return &projectIterator{source: srcIter, fields: t.fields}, nil
}

type projectIterator struct {
	source      database.RowIterator
	fields      []Field
	currentRow  database.Row
	pendingRows []database.Row
}

func (it *projectIterator) Next() bool {
	// 1. Check if we have pending rows from significant unwinding
	if len(it.pendingRows) > 0 {
		it.currentRow = it.pendingRows[0]
		it.pendingRows = it.pendingRows[1:]
		return true
	}

	// 2. Fetch corresponding next row from source
	if it.source.Next() {
		srcRow := it.source.Row()

		// Temporary storage for values (indexed by field index to preserve order later if needed,
		// but we can just iterate fields to build ordered map directly)
		// Wait, unwinding logic needs to know which fields are arrays.
		// So we collect values first.

		type fieldVal struct {
			key      string
			val      interface{}
			isArray  bool
			arrayVal []interface{}
		}

		fVals := make([]fieldVal, len(it.fields))

		allArraysLength := -1
		consistentArrays := true
		hasArrays := false

		for i, f := range it.fields {
			key := f.Alias
			if key == "" {
				key = f.Path
			}

			val, err := srcRow.Get(f.Path)
			if err != nil {
				// Field missing? nil
				val = nil
			}

			fv := fieldVal{key: key, val: val}

			if sliceVal, ok := val.([]interface{}); ok {
				fv.isArray = true
				fv.arrayVal = sliceVal
				hasArrays = true

				if allArraysLength == -1 {
					allArraysLength = len(sliceVal)
				} else if allArraysLength != len(sliceVal) {
					consistentArrays = false
				}
			}
			fVals[i] = fv
		}

		// 3. Unwind Logic
		if hasArrays && consistentArrays && allArraysLength > 0 {
			// Generate N rows
			for i := 0; i < allArraysLength; i++ {
				// Build OrderedMap
				newRow := make(database.OrderedMap, len(it.fields))
				for j, fv := range fVals {
					var v interface{}
					if fv.isArray {
						v = fv.arrayVal[i]
					} else {
						v = fv.val
					}
					newRow[j] = database.KeyVal{Key: fv.key, Val: v}
				}
				it.pendingRows = append(it.pendingRows, database.NewJSONRow(newRow))
			}

			it.currentRow = it.pendingRows[0]
			it.pendingRows = it.pendingRows[1:]
			return true
		}

		// 4. Fallback: Return as is (columnar / mismatched length / scalar only)
		newRow := make(database.OrderedMap, len(it.fields))
		for i, fv := range fVals {
			newRow[i] = database.KeyVal{Key: fv.key, Val: fv.val}
		}
		it.currentRow = database.NewJSONRow(newRow)
		return true
	}
	return false
}

func (it *projectIterator) Row() database.Row {
	return it.currentRow
}

func (it *projectIterator) Error() error {
	return it.source.Error()
}

func (it *projectIterator) Close() error {
	return it.source.Close()
}
