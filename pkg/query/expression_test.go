package query

import (
	"testing"

	"github.com/bisegni/jsl/pkg/parser"
)

func TestBooleanLogic(t *testing.T) {
	record := parser.Record{
		"val":    float64(15),
		"status": "active",
		"type":   "normal",
	}

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "Simple AND - True",
			query:    "SELECT * WHERE val > 10 AND status = 'active'",
			expected: true,
		},
		{
			name:     "Simple AND - False",
			query:    "SELECT * WHERE val > 20 AND status = 'active'",
			expected: false,
		},
		{
			name:     "Simple OR - True",
			query:    "SELECT * WHERE val > 20 OR status = 'active'",
			expected: true,
		},
		{
			name:     "Simple OR - False",
			query:    "SELECT * WHERE val > 20 OR status = 'inactive'",
			expected: false,
		},
		{
			name: "AND with OR - Precedence AND > OR",
			// (True AND False) OR True => False OR True => True
			query:    "SELECT * WHERE val > 10 AND status = 'inactive' OR type = 'normal'",
			expected: true,
		},
		{
			name: "AND with OR - Precedence AND > OR (Case 2)",
			// True OR (False AND True) => True OR False => True
			query:    "SELECT * WHERE val > 10 OR status = 'inactive' AND type = 'error'",
			expected: true,
		},
		{
			name: "Nested Logic (documented example)",
			// (val > 10 AND status = 'active') OR (type = 'critical')
			// (True AND True) OR (False) => True OR False => True
			query:    "SELECT * WHERE (val > 10 AND status = 'active') OR type = 'critical'",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := ParseQuery(tt.query)
			if err != nil {
				t.Fatalf("ParseQuery failed: %v", err)
			}

			if q.Filter == nil {
				t.Fatalf("Expected Filter to be populated")
			}

			result := q.Filter.Evaluate(record)
			if result != tt.expected {
				t.Errorf("Evaluate() = %v, want %v", result, tt.expected)
			}
		})
	}
}
