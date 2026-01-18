package engine

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
)

// Executor runs a Query against an Input Table
type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(q *Query, input database.Table, w io.Writer) error {
	// Refine query to handle $ operator
	if q.Condition != "" {
		refineQuery(q)
	}

	var currentTable database.Table = input

	// Apply WHERE (Filter)
	if q.Condition != "" {
		expr := query.ParseFilterExpression(q.Condition)
		if expr == nil {
			return fmt.Errorf("invalid filter expression: %s", q.Condition)
		}
		currentTable = &FilterTable{
			source: input,
			filter: query.NewFilter(expr.Field, expr.Operator, expr.Value),
		}
	}

	// Apply SELECT (Projection)
	if len(q.Fields) > 0 {
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

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	// Collect all results to print as a JSON array (like original JSL behavior)
	var results []interface{}

	for iterator.Next() {
		results = append(results, iterator.Row().Primitive())
	}

	if err := iterator.Error(); err != nil {
		return err
	}

	return encoder.Encode(results)
}

// FilterTable wraps a source table and filters rows
type FilterTable struct {
	source database.Table
	filter *query.Filter
}

func (t *FilterTable) Iterate() (database.RowIterator, error) {
	srcIter, err := t.source.Iterate()
	if err != nil {
		return nil, err
	}
	return &filterIterator{source: srcIter, filter: t.filter}, nil
}

type filterIterator struct {
	source database.RowIterator
	filter *query.Filter
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

		if it.filter.Match(record) {
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
	source     database.RowIterator
	fields     []Field
	currentRow database.Row
}

func (it *projectIterator) Next() bool {
	if it.source.Next() {
		srcRow := it.source.Row()
		// Construct new projected row
		newMap := make(map[string]interface{})

		for _, f := range it.fields {
			val, err := srcRow.Get(f.Path)
			if err == nil {
				// Use alias if present, otherwise use path
				key := f.Alias
				if key == "" {
					key = f.Path
				}
				newMap[key] = val
			}
		}
		it.currentRow = database.NewJSONRow(newMap)
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
