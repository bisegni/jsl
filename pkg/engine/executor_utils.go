package engine

import (
	"strings"
)

// refineQuery rewrites paths containing $ to use the filter from the WHERE clause
// Example: SELECT sensors.$.name WHERE sensors.*.type='temp' -> SELECT sensors.*.type='temp'.name
func refineQuery(q *Query) {
	// 1. Parse the condition to find the array filter
	// We are looking for something like "ArrayPath.*.FilterSuffix"
	// For now, simple string splitting similar to Extract logic

	// This logic duplicates some of query.go/executor parsing but we need strictly the structure
	// Let's assume the Condition follows the "standard" array filter pattern we support:
	// "path.to.array.*.condition"

	parts := strings.Split(q.Condition, ".*.")
	if len(parts) != 2 {
		return
	}

	arrayPath := parts[0]
	filterSuffix := parts[1]

	// 2. Iterate fields and replace $
	for i := range q.Fields {
		f := &q.Fields[i]
		if strings.Contains(f.Path, "$") {
			// Check if field path starts with array path
			// e.g. "sensors.$.name" vs "sensors"
			// The $ should appear right after the array path
			// Construct expected prefix: "sensors.$"
			expectedPrefix := arrayPath + ".$"

			if strings.HasPrefix(f.Path, expectedPrefix) {
				// Replace $ with *.filterSuffix
				// "sensors.$.name" -> "sensors" + "." + "*" + "." + "filterSuffix" + ".name"
				// actually just replace ".$" with ".*.filterSuffix"
				// Note: filterSuffix might contain operators, but here it's just a string replacement

				replacement := ".*." + filterSuffix
				newPath := strings.Replace(f.Path, ".$", replacement, 1)
				f.Path = newPath
			}
		}
	}
}
