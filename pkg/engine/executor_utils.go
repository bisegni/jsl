package engine

import (
	"fmt"
	"strings"

	"github.com/bisegni/jsl/pkg/query"
)

// refineQuery rewrites paths containing $ to use the filter from the WHERE clause
// Example: SELECT sensors.$.name WHERE sensors.type='temp' -> SELECT sensors.*.type='temp'.name
func refineQuery(q *Query) {
	if q.Condition == "" {
		return
	}

	// Parse the condition to AST
	expr := query.ParseExpression(q.Condition)

	for i := range q.Fields {
		f := &q.Fields[i]
		if strings.Contains(f.Path, "$") {
			// Identify array path: part before .$
			// e.g. sensors.$.name -> arrayPath="sensors"
			dollarIndex := strings.Index(f.Path, ".$")
			if dollarIndex == -1 {
				continue
			}
			arrayPath := f.Path[:dollarIndex]

			// Find a filter in the expression starting with implicit or explicit path
			// Implicit: sensors.type ('sensors' is arrayPath)
			// Explicit: sensors.*.type ('sensors.*' is arrayPath + ".?*?")
			filter := findFilterForArray(expr, arrayPath)
			if filter != nil {
				// Construct filter suffix
				// Field: sensors.type -> Suffix: type
				// Field: sensors.*.type -> Suffix: type
				filterField := filter.Field
				var suffix string
				if strings.HasPrefix(filterField, arrayPath+".*.") {
					suffix = filterField[len(arrayPath)+3:]
				} else if strings.HasPrefix(filterField, arrayPath+".") {
					suffix = filterField[len(arrayPath)+1:]
				} else if filterField == arrayPath {
					// Filter on the array itself? e.g. tags='important' where tags is strings
					// suffix is empty? value comparison
					suffix = "" // special handling needed?
				}

				// Reconstruct replacement: .*.suffix=value
				// Value needs to be handled (quotes etc).
				// Filter.Value is interface{}.
				valStr := fmt.Sprintf("%v", filter.Value)
				// Quote string values if needed (simple heuristic)
				if _, ok := filter.Value.(string); ok {
					valStr = "'" + valStr + "'"
				}

				replacement := ".*"
				if suffix != "" {
					replacement += "." + suffix
				}
				replacement += filter.Operator + valStr

				// Replace .$ with replacement
				// replacement is ".*.type='temp'"
				// f.Path was "sensors.$.name" -> "sensors" + replacement + ".name"

				// Wait, "sensors" + ".*.type='temp'" + ".name" -> "sensors.*.type='temp'.name"
				// Looks correct.

				// Update path
				// We replace ".$"
				f.Path = strings.Replace(f.Path, ".$", replacement, 1)
			}
		}
	}
}

func findFilterForArray(expr query.Expression, arrayPath string) *query.Filter {
	// BFS or DFS traversal
	switch e := expr.(type) {
	case *query.Condition:
		// Check if field belongs to arrayPath
		// 1. Implicit: field starts with "arrayPath."
		// 2. Explicit: field starts with "arrayPath.*."
		if strings.HasPrefix(e.Filter.Field, arrayPath+".") {
			return e.Filter
		}
	case *query.AndExpression:
		// Try left, then right
		if f := findFilterForArray(e.Left, arrayPath); f != nil {
			return f
		}
		return findFilterForArray(e.Right, arrayPath)
	case *query.OrExpression:
		// Ambiguous for OR. Return first match?
		if f := findFilterForArray(e.Left, arrayPath); f != nil {
			return f
		}
		return findFilterForArray(e.Right, arrayPath)
	}
	return nil
}
