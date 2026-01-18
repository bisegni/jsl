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

	// Smart split: split by dots, but preserve dots inside filter expressions
	// A dot is a separator IF it's not followed by an operator before the next dot
	operators := []string{">=", "<=", "!=", "~=", ">", "<", "="}
	var parts []string
	var current strings.Builder

	for i := 0; i < len(path); i++ {
		if path[i] == '.' {
			// Check if this dot is a separator
			// Look ahead for an operator before the next dot
			isSeparator := true
			rest := path[i+1:]
			nextDot := strings.Index(rest, ".")
			segment := rest
			if nextDot != -1 {
				segment = rest[:nextDot]
			}

			for _, op := range operators {
				if strings.Contains(segment, op) {
					// Exception: if the previous part was a wildcard, we MUST split here
					// regardless of operators.
					// e.g. "foo.*.value>20" -> "foo", "*", "value>20"
					// If we don't split, we get "foo", "*.value>20" which is wrong.
					if current.String() == "*" || current.String() == "%" {
						isSeparator = true
					} else {
						isSeparator = false
					}
					break
				}
			}

			if isSeparator {
				parts = append(parts, current.String())
				current.Reset()
				continue
			}
		}
		current.WriteByte(path[i])
	}
	parts = append(parts, current.String())

	// Filter out empty parts
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// extractFromMap handles extracting values from a map, supporting wildcards and operators
func extractFromMap(m map[string]interface{}, part string, remaining []string) (interface{}, error) {
	// Check if this part is a filter expression (e.g., "type=temp")
	if IsFilterExpression(part) {
		expr := ParseFilterExpression(part)
		if expr != nil {
			// Extract the field from the current map to check the condition
			q := NewQuery(expr.Field)
			val, err := q.Extract(m)
			if err == nil {
				// We found the field, now compare
				// Parse filter value for comparison (try number first)
				var filterVal interface{}
				filterVal = expr.Value
				if n, err := strconv.ParseFloat(expr.Value, 64); err == nil {
					filterVal = n
				}

				match := false
				switch expr.Operator {
				case "=", "==":
					match = compareEqual(val, filterVal)
				case "!=":
					match = !compareEqual(val, filterVal)
				case ">":
					match = compareGreater(val, filterVal)
				case ">=":
					match = compareGreaterEqual(val, filterVal)
				case "<":
					match = compareLess(val, filterVal)
				case "<=":
					match = compareLessEqual(val, filterVal)
				case "contains":
					match = containsValue(val, filterVal)
				}

				if match {
					// Condition met! Continue with remaining path on the SAME map
					return extractValue(m, remaining)
				}
				return nil, fmt.Errorf("filter '%s' did not match", part)
			}
		}
	}

	// Simple key access
	if !strings.HasPrefix(part, "*") && !strings.HasPrefix(part, "%") {
		if val, ok := m[part]; ok {
			return extractValue(val, remaining)
		}
		return nil, fmt.Errorf("key '%s' not found", part)
	}

	// Wildcard access
	var operator string
	var filterValue string

	if part == "*" || part == "%" {
		operator = "*" // match all
	} else {
		// Try to find an operator
		operators := []string{">=", "<=", "!=", "~=", ">", "<", "="}
		wildcards := []string{"*", "%"}
		for _, w := range wildcards {
			for _, op := range operators {
				if strings.HasPrefix(part, w+op) {
					operator = op
					filterValue = part[len(op)+1:]
					goto found
				}
			}
		}
	found:
		if operator == "" {
			return nil, fmt.Errorf("invalid wildcard filter: %s", part)
		}
	}

	results := make(map[string]interface{})
	for k, v := range m {
		match := false
		switch operator {
		case "*":
			match = true
		case "=":
			match = k == filterValue
		case "!=":
			match = k != filterValue
		case "~=":
			match = strings.Contains(k, filterValue)
		case ">":
			match = k > filterValue
		case ">=":
			match = k >= filterValue
		case "<":
			match = k < filterValue
		case "<=":
			match = k <= filterValue
		}

		if match {
			val, err := extractValue(v, remaining)
			if err == nil {
				results[k] = val
			}
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no keys matched wildcard filter '%s'", part)
	}
	return results, nil
}

func extractValue(data interface{}, parts []string) (interface{}, error) {
	if len(parts) == 0 {
		return data, nil
	}

	part := parts[0]
	remaining := parts[1:]

	switch v := data.(type) {
	case parser.Record:
		// Handle parser.Record (which is map[string]interface{})
		return extractFromMap(v, part, remaining)

	case map[string]interface{}:
		// Handle object access
		return extractFromMap(v, part, remaining)

	case []interface{}:
		// Handle array access
		if part == "*" || part == "%" {
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

	return f.matchValue(value)
}

func (f *Filter) matchValue(value interface{}) bool {
	// Handle collections - if ANY element matches, the filter matches
	switch v := value.(type) {
	case map[string]interface{}:
		for _, val := range v {
			if f.matchValue(val) {
				return true
			}
		}
		return false
	case []interface{}:
		for _, val := range v {
			if f.matchValue(val) {
				return true
			}
		}
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
	// Try direct comparison for common types
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	}
	// Fallback to string comparison for other types
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
	// Handle string types directly for efficiency
	if aStr, ok := a.(string); ok {
		if bStr, ok := b.(string); ok {
			return strings.Contains(aStr, bStr)
		}
		// If b is not a string, convert it
		bStr := fmt.Sprintf("%v", b)
		return strings.Contains(aStr, bStr)
	}
	// Fallback to string conversion for other types
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

// FilterExpr represents a parsed filter expression
type FilterExpr struct {
	Field    string
	Operator string
	Value    string
}

// IsFilterExpression checks if a string looks like a filter expression (contains an operator)
// and does NOT start with a dot (which signifies a path query)
func IsFilterExpression(expr string) bool {
	if strings.HasPrefix(expr, ".") {
		return false
	}
	operators := []string{">=", "<=", "!=", "~=", ">", "<", "="}
	for _, op := range operators {
		if strings.Contains(expr, op) {
			return true
		}
	}
	return false
}

// ParseFilterExpression parses expressions like "age>28", "name=john", "status!=active"
func ParseFilterExpression(expr string) *FilterExpr {
	// Try to find operator in the expression
	operators := []string{">=", "<=", "!=", "~=", ">", "<", "="}

	for _, op := range operators {
		if idx := strings.Index(expr, op); idx > 0 {
			field := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+len(op):])

			if field != "" && value != "" {
				// Convert ~= to contains for internal representation
				internalOp := op
				if op == "~=" {
					internalOp = "contains"
				}
				return &FilterExpr{
					Field:    field,
					Operator: internalOp,
					Value:    value,
				}
			}
		}
	}

	return nil
}
