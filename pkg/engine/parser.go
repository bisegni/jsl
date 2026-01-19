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
	Condition string
	GroupBy   string
}

// ParseQuery parses a SELECT string.
// Syntax: SELECT <fields> [WHERE <condition>] [GROUP BY <field>]
// Example: SELECT room, AVG(val) AS avg_val WHERE val > 0 GROUP BY room
func ParseQuery(input string) (*Query, error) {
	input = strings.TrimSpace(input)

	// Case-insensitive check for SELECT
	if len(input) < 6 || !strings.EqualFold(input[:6], "SELECT") {
		return nil, fmt.Errorf("query must start with SELECT")
	}

	rest := input[6:]

	// Check for GROUP BY (must be checked first if it appears after WHERE, or maybe split by keywords)
	// Simplest: Find GROUP BY, then WHERE in the remaining part.
	// But WHERE usually comes before GROUP BY.

	groupByIndex := -1
	upper := strings.ToUpper(rest)
	if idx := strings.Index(upper, " GROUP BY "); idx != -1 {
		groupByIndex = idx
	}

	var groupBy string
	if groupByIndex != -1 {
		groupBy = strings.TrimSpace(rest[groupByIndex+10:]) // 10 = len(" GROUP BY ")
		rest = rest[:groupByIndex]
		upper = upper[:groupByIndex] // truncate upper for WHERE search
	}

	// Check for WHERE clause
	whereIndex := -1
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
		Condition: condition,
		GroupBy:   groupBy,
	}, nil
}
