package engine

import (
	"fmt"
	"strings"
)

// Field represents a selected field with optional alias
type Field struct {
	Path  string
	Alias string
}

// Query represents a parsed SQL-like query
type Query struct {
	Fields    []Field
	Condition string
}

// ParseQuery parses a SELECT string.
// Syntax: SELECT <fields> [WHERE <condition>]
// Example: SELECT name, age AS user_age WHERE age > 25
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

	var fields []Field
	if fieldsStr == "*" || fieldsStr == "" {
		fields = []Field{} // Empty means all/wildcard
	} else {
		parts := strings.Split(fieldsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				// Check for AS alias
				var path, alias string
				pUpper := strings.ToUpper(p)
				asIndex := strings.LastIndex(pUpper, " AS ")
				if asIndex != -1 {
					path = strings.TrimSpace(p[:asIndex])
					alias = strings.TrimSpace(p[asIndex+4:])
				} else {
					path = p
					alias = p
				}
				fields = append(fields, Field{Path: path, Alias: alias})
			}
		}
	}

	return &Query{
		Fields:    fields,
		Condition: condition,
	}, nil
}
