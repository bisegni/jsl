package engine

import (
	"encoding/json"
	"io"

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
	// Refine query to handle $ operator
	if q.Condition != "" {
		refineQuery(q)
	}

	var currentTable database.Table = input

	// Apply WHERE (Filter)
	if q.Condition != "" {
		expr := query.ParseExpression(q.Condition)
		currentTable = &FilterTable{
			source:     input,
			expression: expr,
		}
	}

	// Apply SELECT (Projection) or Aggregation
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

	// Iterate and Print Results
	iterator, err := currentTable.Iterate()
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

		// Temporary map to hold values before determining projection strategy
		rowMap := make(map[string]interface{})

		// Track array properties for potential unwinding
		type arrayField struct {
			key string
			val []interface{}
		}
		var arrayFields []arrayField
		var scalarFields []string // keys of scalar fields

		hasArrays := false
		allArraysLength := -1
		consistentArrays := true

		for _, f := range it.fields {
			val, err := srcRow.Get(f.Path)
			if err == nil {
				key := f.Alias
				if key == "" {
					key = f.Path
				}
				rowMap[key] = val

				// Check if value is slice
				if sliceVal, ok := val.([]interface{}); ok {
					hasArrays = true
					if allArraysLength == -1 {
						allArraysLength = len(sliceVal)
					} else if allArraysLength != len(sliceVal) {
						consistentArrays = false
					}
					arrayFields = append(arrayFields, arrayField{key: key, val: sliceVal})
				} else {
					scalarFields = append(scalarFields, key)
				}
			}
		}

		// 3. Unwind Logic
		// If we have arrays, they are consistent in length, and length > 0, we unwind/zip them.
		if hasArrays && consistentArrays && allArraysLength > 0 {
			// Generate N rows
			for i := 0; i < allArraysLength; i++ {
				newMap := make(map[string]interface{})

				// fill arrays
				for _, af := range arrayFields {
					newMap[af.key] = af.val[i]
				}
				// fill scalars (repeat)
				for _, k := range scalarFields {
					newMap[k] = rowMap[k]
				}

				it.pendingRows = append(it.pendingRows, database.NewJSONRow(newMap))
			}

			// Return the first one
			it.currentRow = it.pendingRows[0]
			it.pendingRows = it.pendingRows[1:]
			return true
		}

		// 4. Fallback: Return as is (columnar / mismatched length / scalar only)
		it.currentRow = database.NewJSONRow(rowMap)
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
