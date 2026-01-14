package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bisegni/jsl/pkg/parser"
)

// Query represents a path-based query
type Query struct {
	Path string
}

// NewQuery creates a new query from a path string
func NewQuery(path string) *Query {
	return &Query{Path: path}
}

// Extract extracts values from a record using the path
func (q *Query) Extract(record parser.Record) (interface{}, error) {
	if q.Path == "" || q.Path == "." {
		return record, nil
	}

	parts := parsePath(q.Path)
	return extractValue(record, parts)
}

// parsePath parses a dot-separated path into parts
func parsePath(path string) []string {
	// Remove leading dot if present
	path = strings.TrimPrefix(path, ".")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ".")
}

// extractValue extracts a value from nested maps/arrays
func extractValue(data interface{}, parts []string) (interface{}, error) {
	if len(parts) == 0 {
		return data, nil
	}

	part := parts[0]
	remaining := parts[1:]

	switch v := data.(type) {
	case parser.Record:
		// Handle parser.Record (which is map[string]interface{})
		if val, ok := v[part]; ok {
			return extractValue(val, remaining)
		}
		return nil, fmt.Errorf("key '%s' not found", part)
	
	case map[string]interface{}:
		// Handle object access
		if val, ok := v[part]; ok {
			return extractValue(val, remaining)
		}
		return nil, fmt.Errorf("key '%s' not found", part)
	
	case []interface{}:
		// Handle array access
		if part == "*" {
			// Wildcard - extract from all elements
			results := make([]interface{}, 0, len(v))
			for _, item := range v {
				val, err := extractValue(item, remaining)
				if err == nil {
					results = append(results, val)
				}
			}
			return results, nil
		}
		
		// Numeric index
		idx, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid array index '%s'", part)
		}
		if idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("array index %d out of bounds", idx)
		}
		return extractValue(v[idx], remaining)
	
	default:
		return nil, fmt.Errorf("cannot access '%s' on type %T", part, data)
	}
}

// Filter represents a filtering condition
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// NewFilter creates a new filter
func NewFilter(field, operator string, value interface{}) *Filter {
	return &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// Match checks if a record matches the filter
func (f *Filter) Match(record parser.Record) bool {
	q := NewQuery(f.Field)
	value, err := q.Extract(record)
	if err != nil {
		return false
	}

	switch f.Operator {
	case "=", "==":
		return compareEqual(value, f.Value)
	case "!=":
		return !compareEqual(value, f.Value)
	case ">":
		return compareGreater(value, f.Value)
	case ">=":
		return compareGreaterEqual(value, f.Value)
	case "<":
		return compareLess(value, f.Value)
	case "<=":
		return compareLessEqual(value, f.Value)
	case "contains":
		return containsValue(value, f.Value)
	default:
		return false
	}
}

func compareEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func compareGreater(a, b interface{}) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af > bf
	}
	return false
}

func compareGreaterEqual(a, b interface{}) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af >= bf
	}
	return false
}

func compareLess(a, b interface{}) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af < bf
	}
	return false
}

func compareLessEqual(a, b interface{}) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af <= bf
	}
	return false
}

func containsValue(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Contains(aStr, bStr)
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		return f, err == nil
	}
}
