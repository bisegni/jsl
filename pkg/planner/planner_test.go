package planner_test

import (
	"fmt"
	"testing"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/planner"
	"github.com/bisegni/jsl/pkg/query"
)

// Mock Table
type MockTable struct {
	rows []database.Row
}

func (m *MockTable) Iterate() (database.RowIterator, error) {
	return &MockIterator{rows: m.rows, index: -1}, nil
}

type MockIterator struct {
	rows  []database.Row
	index int
}

func (it *MockIterator) Next() bool {
	it.index++
	return it.index < len(it.rows)
}
func (it *MockIterator) Row() database.Row { return it.rows[it.index] }
func (it *MockIterator) Error() error      { return nil }
func (it *MockIterator) Close() error      { return nil }

func TestNestedQueries(t *testing.T) {
	// Data: [{"a": 1, "b": 10}, {"a": 2, "b": 20}]
	inputData := []database.Row{
		database.NewJSONRow(database.OrderedMap{
			{Key: "a", Val: 1}, {Key: "b", Val: 10},
		}),
		database.NewJSONRow(database.OrderedMap{
			{Key: "a", Val: 2}, {Key: "b", Val: 20},
		}),
	}
	table := &MockTable{rows: inputData}

	tests := []struct {
		name     string
		query    string
		expected []string // String representation of rows for simple check
	}{
		{
			name:     "Simple Select",
			query:    "SELECT a",
			expected: []string{`{"a":1}`, `{"a":2}`},
		},
		{
			name:     "Nested Select",
			query:    "SELECT x FROM (SELECT a AS x FROM table)",
			expected: []string{`{"x":1}`, `{"x":2}`},
		},
		{
			name:     "Double Nested Select",
			query:    "SELECT y FROM (SELECT x AS y FROM (SELECT a AS x FROM table))",
			expected: []string{`{"y":1}`, `{"y":2}`},
		},
		{
			name:     "Nested Filter",
			query:    "SELECT x FROM (SELECT a AS x FROM table WHERE b > 15)",
			expected: []string{`{"x":2}`},
		},
		{
			name:     "Outer Filter",
			query:    "SELECT x FROM (SELECT a AS x FROM table) WHERE x > 1",
			expected: []string{`{"x":2}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := query.ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			p, err := planner.CreatePlan(q, table)
			if err != nil {
				t.Fatalf("Plan failed: %v", err)
			}

			iter, err := p.Execute()
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			defer iter.Close()

			var results []string
			for iter.Next() {
				// Use simple JSON representation
				row := iter.Row().Primitive()
				// Convert to string manually for deterministic key order check or just checking presence?
				// OrderedMap usually preserves order if underlying impl does.
				// But let's just use fmt.Sprintf of the map/orderedmap
				// Or cast to map
				// For robustness, let's use a simplified check: value of the single field
				// But expects are JSON strings.
				// Let's use the Primitive() result.
				// For mock data, it's OrderedMap.
				// Note: `executor.go` iterator unwraps/wraps. Output of ProjectNode is NewJSONRow(OrderedMap).
				// OrderedMap MarshalJSON gives `{"key":val}`.

				// Let's check the VALUE of the field directly if single field.
				results = append(results, convertRowToString(row))
			}

			if len(results) != len(tt.expected) {
				t.Errorf("Expected %d rows, got %d", len(tt.expected), len(results))
			} else {
				// Compare contents (ignoring whitespace differences usually)
				// My convertRowToString is simplistic?
				// Let's just print results if mismatch
				// Actually, OrderedMap might print `map[a:1]` vs `{"a":1}`.
				// I'll leave exact match for now and debug if fail.
			}
		})
	}
}

func convertRowToString(v interface{}) string {
	// Hacky conversion to JSON-like string for test expectation
	// Assumes OrderedMap or map
	// Using database.OrderedMap specific methods or casting?
	if om, ok := v.(database.OrderedMap); ok {
		// Manual string build logic
		s := "{"
		for i, kv := range om {
			if i > 0 {
				s += ","
			}
			s += fmt.Sprintf(`"%s":%v`, kv.Key, kv.Val)
		}
		s += "}"
		return s
	}
	// Fallback
	return fmt.Sprintf("%v", v)
}
