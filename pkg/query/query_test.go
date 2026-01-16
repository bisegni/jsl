package query

import (
	"testing"

	"github.com/bisegni/jsl/pkg/parser"
)

func TestQueryExtract(t *testing.T) {
	record := parser.Record{
		"name": "Alice",
		"age":  float64(30),
		"address": map[string]interface{}{
			"city":  "New York",
			"state": "NY",
		},
		"tags": []interface{}{"golang", "testing", "json"},
	}

	tests := []struct {
		name     string
		path     string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple field",
			path:     "name",
			expected: "Alice",
			wantErr:  false,
		},
		{
			name:     "numeric field",
			path:     "age",
			expected: float64(30),
			wantErr:  false,
		},
		{
			name:     "nested field",
			path:     "address.city",
			expected: "New York",
			wantErr:  false,
		},
		{
			name:     "array element",
			path:     "tags.0",
			expected: "golang",
			wantErr:  false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: nil, // Will be checked separately
			wantErr:  false,
		},
		{
			name:    "non-existent field",
			path:    "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery(tt.path)
			result, err := q.Extract(record)

			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Special case for empty path - returns the whole record
				if tt.path == "" || tt.path == "." {
					if _, ok := result.(parser.Record); !ok {
						t.Errorf("Extract() with empty path should return parser.Record, got %T", result)
					}
				} else if result != tt.expected {
					t.Errorf("Extract() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestFilterMatch(t *testing.T) {
	record := parser.Record{
		"name": "Alice",
		"age":  float64(30),
		"city": "New York",
	}

	tests := []struct {
		name     string
		field    string
		operator string
		value    interface{}
		expected bool
	}{
		{
			name:     "equal string",
			field:    "name",
			operator: "=",
			value:    "Alice",
			expected: true,
		},
		{
			name:     "not equal string",
			field:    "name",
			operator: "!=",
			value:    "Bob",
			expected: true,
		},
		{
			name:     "greater than",
			field:    "age",
			operator: ">",
			value:    float64(25),
			expected: true,
		},
		{
			name:     "less than",
			field:    "age",
			operator: "<",
			value:    float64(35),
			expected: true,
		},
		{
			name:     "greater than or equal",
			field:    "age",
			operator: ">=",
			value:    float64(30),
			expected: true,
		},
		{
			name:     "contains",
			field:    "city",
			operator: "contains",
			value:    "York",
			expected: true,
		},
		{
			name:     "does not contain",
			field:    "city",
			operator: "contains",
			value:    "Boston",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFilter(tt.field, tt.operator, tt.value)
			result := f.Match(record)

			if result != tt.expected {
				t.Errorf("Match() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWildcardExtract(t *testing.T) {
	record := parser.Record{
		"employees": []interface{}{
			map[string]interface{}{"name": "John", "role": "Engineer"},
			map[string]interface{}{"name": "Jane", "role": "Manager"},
		},
	}

	q := NewQuery("employees.*.name")
	result, err := q.Extract(record)
	if err != nil {
		t.Fatalf("Extract() with * failed: %v", err)
	}

	names, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	// Test shell-safe wildcard %
	q = NewQuery("employees.%.name")
	result, err = q.Extract(record)
	if err != nil {
		t.Fatalf("Extract() with %% failed: %v", err)
	}

	names, ok = result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
}
func TestWildcardKeyExtract(t *testing.T) {
	record := parser.Record{
		"metrics": map[string]interface{}{
			"temp_input":  float64(25.5),
			"temp_output": float64(26.0),
			"humidity":    float64(45),
			"pressure":    float64(1013),
		},
	}

	tests := []struct {
		name     string
		path     string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "wildcard match all",
			path: "metrics.*",
			expected: map[string]interface{}{
				"temp_input":  float64(25.5),
				"temp_output": float64(26.0),
				"humidity":    float64(45),
				"pressure":    float64(1013),
			},
			wantErr: false,
		},
		{
			name: "wildcard match contains",
			path: "metrics.*~=temp",
			expected: map[string]interface{}{
				"temp_input":  float64(25.5),
				"temp_output": float64(26.0),
			},
			wantErr: false,
		},
		{
			name: "wildcard match equals",
			path: "metrics.*=humidity",
			expected: map[string]interface{}{
				"humidity": float64(45),
			},
			wantErr: false,
		},
		{
			name: "wildcard match not equals",
			path: "metrics.*!=humidity",
			expected: map[string]interface{}{
				"temp_input":  float64(25.5),
				"temp_output": float64(26.0),
				"pressure":    float64(1013),
			},
			wantErr: false,
		},
		{
			name: "wildcard match greater equal",
			path: "metrics.*>=pressure",
			expected: map[string]interface{}{
				"pressure":    float64(1013),
				"temp_input":  float64(25.5),
				"temp_output": float64(26.0),
			},
			wantErr: false,
		},
		{
			name: "shell-safe wildcard match contains",
			path: "metrics.%~=temp",
			expected: map[string]interface{}{
				"temp_input":  float64(25.5),
				"temp_output": float64(26.0),
			},
			wantErr: false,
		},
		{
			name:    "no match",
			path:    "metrics.*=nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery(tt.path)
			result, err := q.Extract(record)

			if (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected map[string]interface{}, got %T", result)
				}
				if len(got) != len(tt.expected) {
					t.Errorf("Expected %d results, got %d", len(tt.expected), len(got))
				}
				for k, v := range tt.expected {
					if got[k] != v {
						t.Errorf("For key %s, expected %v, got %v", k, v, got[k])
					}
				}
			}
		})
	}
}
