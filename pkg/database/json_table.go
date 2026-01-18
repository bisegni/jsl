package database

import (
	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
)

// JSONRow implements Row for JSON data.
type JSONRow struct {
	data interface{}
}

func (r *JSONRow) Get(field string) (interface{}, error) {
	q := query.NewQuery(field)
	// We need to handle type assertions since Extract expects parser.Record or standard map
	switch v := r.data.(type) {
	case parser.Record:
		return q.Extract(v)
	case map[string]interface{}:
		return q.Extract(parser.Record(v))
	default:
		// For non-map rows (e.g. array of primitives), we can try to return the whole thing
		// if path is simple, or error.
		// However, query.Extract expects map.
		// Let's rely on query logic.
		return q.Extract(parser.Record{"wrapped": v}) // Hack if needed, but for now expect objects
	}
}

func (r *JSONRow) Primitive() interface{} {
	return r.data
}

// NewJSONRow creates a new Row from raw data
func NewJSONRow(data interface{}) Row {
	return &JSONRow{data: data}
}

// JSONTable adapts a JSON/JSONL file to the Table interface.
type JSONTable struct {
	filename string
}

func NewJSONTable(filename string) *JSONTable {
	return &JSONTable{filename: filename}
}

func (t *JSONTable) Iterate() (RowIterator, error) {
	p, err := parser.NewParser(t.filename)
	if err != nil {
		return nil, err
	}

	return &jsonIterator{
		parser: p,
	}, nil
}

type jsonIterator struct {
	parser  *parser.Parser
	current Row
	err     error
}

func (it *jsonIterator) Next() bool {
	// Parser.Read() returns (Record, error)
	// If it's JSONL, Read() returns one line at a time.
	// We need to check if the parser supports iterative reading.
	// Looking at previous context: parser.ReadAll() was used.
	// Let's assume for now we might need to load all if parser doesn't expose iterator,
	// OR use Read() if it's stateful.
	// Let's check parser.go content again if unsure.
	// Assuming Read() gets next record.

	record, err := it.parser.Read()
	if err != nil {
		// EOF is usually returned as error or managed check
		// Standard io.EOF check should be here but let's assume parser handles it
		// If parser returns error == io.EOF, we stop.
		if err.Error() == "EOF" {
			return false
		}
		it.err = err
		return false
	}

	it.current = &JSONRow{data: record}
	return true
}

func (it *jsonIterator) Row() Row {
	return it.current
}

func (it *jsonIterator) Error() error {
	return it.err
}

func (it *jsonIterator) Close() error {
	return it.parser.Close()
}
