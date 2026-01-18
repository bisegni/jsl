package engine

import (
	"fmt"
	"strings"
)

// Query represents a parsed SQL-like query
type Query struct {
	Fields    []string
	Condition string
}

// ParseQuery parses a SELECT string.
// Syntax: SELECT <fields> [WHERE <condition>]
// Example: SELECT name, age WHERE age > 25
func ParseQuery(input string) (*Query, error) {
	input = strings.TrimSpace(input)

	// Case-insensitive check for SELECT
	if len(input) < 6 || !strings.EqualFold(input[:6], "SELECT") {
		return nil, fmt.Errorf("query must start with SELECT")
	}

	rest := input[6:]

	// Check for WHERE clause
	whereIndex := -1
	upper := strings.ToUpper(rest)
	if idx := strings.Index(upper, " WHERE "); idx != -1 {
		whereIndex = idx
	}

	var fieldsStr string
	var condition string

	if whereIndex != -1 {
		fieldsStr = rest[:whereIndex]
		condition = strings.TrimSpace(rest[whereIndex+7:])
	} else {
		fieldsStr = rest
	}

	fieldsStr = strings.TrimSpace(fieldsStr)

	var fields []string
	if fieldsStr == "*" || fieldsStr == "" {
		fields = []string{} // Empty means all/wildcard
	} else {
		parts := strings.Split(fieldsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				fields = append(fields, p)
			}
		}
	}

	return &Query{
		Fields:    fields,
		Condition: condition,
	}, nil
}
