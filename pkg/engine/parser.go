package engine

import (
	"fmt"
	"strings"
)

// Field represents a selected field with optional alias and aggregation
type Field struct {
	Path      string
	Alias     string
	Aggregate string // "MAX", "MIN", "AVG", "COUNT", "SUM" or empty
}

// Query represents a parsed SQL-like query
type Query struct {
	Fields    []Field
	From      string // Subquery or source
	Condition string
	GroupBy   string
}

// ParseQuery parses a SELECT string.
// Syntax: SELECT <fields> [FROM <source>] [WHERE <condition>] [GROUP BY <field>]
// Example: SELECT room, AVG(val) AS avg_val FROM (SELECT ...) WHERE val > 0 GROUP BY room
func ParseQuery(input string) (*Query, error) {
	input = strings.TrimSpace(input)

	// Case-insensitive check for SELECT
	if len(input) < 6 || !strings.EqualFold(input[:6], "SELECT") {
		return nil, fmt.Errorf("query must start with SELECT")
	}

	rest := input[6:]

	// Helper to find top-level keywords (ignoring parens)
	findKeyword := func(s string, keyword string) int {
		upper := strings.ToUpper(s)
		key := " " + keyword + " " // ensure word boundary
		depth := 0
		for i := 0; i < len(s); i++ {
			if s[i] == '(' {
				depth++
			} else if s[i] == ')' {
				depth--
			} else if depth == 0 {
				// Check for match
				if i+len(key) <= len(s) {
					if upper[i:i+len(key)] == key {
						return i
					}
				}
			}
		}
		return -1
	}

	// 1. Find GROUP BY (Last clause usually)
	groupByIndex := findKeyword(rest, "GROUP BY")
	var groupBy string
	if groupByIndex != -1 {
		groupBy = strings.TrimSpace(rest[groupByIndex+10:]) // 10 = len(" GROUP BY ")
		// Remove optional parens around group by field
		if strings.HasPrefix(groupBy, "(") && strings.HasSuffix(groupBy, ")") {
			groupBy = strings.TrimSpace(groupBy[1 : len(groupBy)-1])
		}
		rest = rest[:groupByIndex]
	}

	// 2. Find WHERE
	whereIndex := findKeyword(rest, "WHERE")
	var condition string
	if whereIndex != -1 {
		condition = strings.TrimSpace(rest[whereIndex+7:])
		rest = rest[:whereIndex]
	}

	// 3. Find FROM
	fromIndex := findKeyword(rest, "FROM")
	var from string
	if fromIndex != -1 {
		from = strings.TrimSpace(rest[fromIndex+6:])
		// Remove optional parens around subquery if strictly formatted
		// But usually `FROM (SELECT ...)` -> `(SELECT ...)`
		// We can strip them here if it looks like a subquery
		// But let's keep them and let recursive parser handle or strip specific wrapping
		if strings.HasPrefix(from, "(") && strings.HasSuffix(from, ")") {
			from = strings.TrimSpace(from[1 : len(from)-1])
		}
		rest = rest[:fromIndex]
	}

	fieldsStr := strings.TrimSpace(rest)

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

				rawField := p
				if asIndex != -1 {
					rawField = strings.TrimSpace(p[:asIndex])
					alias = strings.TrimSpace(p[asIndex+4:])
				} else {
					alias = "" // derived later or redundant
				}

				// Check for Aggregation Function: FUNC(path)
				var aggregate string

				// List of supported aggregates
				aggs := []string{"MAX", "MIN", "AVG", "COUNT", "SUM"}
				upperRaw := strings.ToUpper(rawField)

				for _, agg := range aggs {
					prefix := agg + "("
					if strings.HasPrefix(upperRaw, prefix) && strings.HasSuffix(upperRaw, ")") {
						aggregate = agg
						// Extract content inside parens
						path = strings.TrimSpace(rawField[len(prefix) : len(rawField)-1])
						break
					}
				}

				if aggregate == "" {
					path = rawField
				}

				// Default alias if empty
				if alias == "" {
					if aggregate != "" {
						alias = fmt.Sprintf("%s_%s", strings.ToLower(aggregate), strings.ReplaceAll(path, ".", "_"))
					} else {
						// e.g. sensors.name -> name? or sensors.name?
						// Parser previously just put p as alias if no AS
						alias = path
					}
				}

				fields = append(fields, Field{
					Path:      path,
					Alias:     alias,
					Aggregate: aggregate,
				})
			}
		}
	}

	return &Query{
		Fields:    fields,
		From:      from,
		Condition: condition,
		GroupBy:   groupBy,
	}, nil
}
