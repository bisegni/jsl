package engine

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bisegni/jsl/pkg/database"
)

// AggregateTable wraps a source table and performs aggregation
type AggregateTable struct {
	source database.Table
	query  *Query // Contains GroupBy and Fields with Aggregates
}

func (t *AggregateTable) Iterate() (database.RowIterator, error) {
	srcIter, err := t.source.Iterate()
	if err != nil {
		return nil, err
	}
	return newAggregatorIterator(srcIter, t.query)
}

type aggregatorIterator struct {
	results []database.Row
	index   int
}

func (it *aggregatorIterator) Next() bool {
	it.index++
	return it.index < len(it.results)
}

func (it *aggregatorIterator) Row() database.Row {
	if it.index >= 0 && it.index < len(it.results) {
		return it.results[it.index]
	}
	return nil
}

func (it *aggregatorIterator) Error() error {
	return nil
}

func (it *aggregatorIterator) Close() error {
	return nil
}

func newAggregatorIterator(source database.RowIterator, q *Query) (*aggregatorIterator, error) {
	// 1. Scan source and map groups
	groups := make(map[string]*groupState)
	var groupKeys []string // preserve order or sort? Sorting is better for deterministic output.

	hasData := false

	// Helper to extract value safely
	// Helper to extract value safely
	extract := func(row database.Row, path string) (interface{}, error) {
		return row.Get(path)
	}

	for source.Next() {
		hasData = true
		row := source.Row()

		// Determine Group Key
		var groupKey string
		if q.GroupBy != "" {
			val, err := extract(row, q.GroupBy)
			if err == nil {
				groupKey = fmt.Sprintf("%v", val)
			} else {
				groupKey = "null"
			}
		} else {
			groupKey = "" // Single group for entire dataset
		}

		state, exists := groups[groupKey]
		if !exists {
			state = newGroupState(q.Fields)
			groups[groupKey] = state
			groupKeys = append(groupKeys, groupKey)
		}

		state.update(row, extract)
	}

	if err := source.Error(); err != nil {
		source.Close()
		return nil, err
	}
	source.Close()

	// 2. Build results
	var results []database.Row

	// Handle empty input with global aggregation (e.g. SELECT COUNT(*) should return 0)
	if !hasData && q.GroupBy == "" && len(q.Fields) > 0 {
		// Only if we have aggregations?
		// "SELECT val" -> empty
		// "SELECT COUNT(val)" -> 0
		// Check if any aggregate function exists
		hasAgg := false
		for _, f := range q.Fields {
			if f.Aggregate != "" {
				hasAgg = true
				break
			}
		}
		if hasAgg {
			// Create a default group state
			state := newGroupState(q.Fields)
			// No update call
			results = append(results, state.finalize("", ""))
			return &aggregatorIterator{results: results, index: -1}, nil
		}
	}

	sort.Strings(groupKeys)

	for _, key := range groupKeys {
		state := groups[key]
		results = append(results, state.finalize(key, q.GroupBy))
	}

	return &aggregatorIterator{results: results, index: -1}, nil
}

type groupState struct {
	fields []Field
	aggs   map[string]fieldAggregator
}

func newGroupState(fields []Field) *groupState {
	s := &groupState{
		fields: fields,
		aggs:   make(map[string]fieldAggregator),
	}
	for i, f := range s.fields {
		if f.Aggregate != "" {
			s.aggs[keyFor(i)] = createAggregator(f.Aggregate)
		}
	}
	return s
}

func keyFor(index int) string {
	return strconv.Itoa(index)
}

func (s *groupState) update(row database.Row, extractor func(database.Row, string) (interface{}, error)) {
	for i, f := range s.fields {
		// If it's an aggregate field, update aggregator
		if f.Aggregate != "" {
			val, err := extractor(row, f.Path)
			if err == nil {
				s.aggs[keyFor(i)].Add(val)
			}
		}
		// If it's a regular field (groupBy key or implicit first value), we don't store it here explicitly
		// We re-extract key at finalize or rely on convention (in SQL, non-agg fields must be in GROUP BY)
	}
}

func (s *groupState) finalize(groupKey string, groupByField string) database.Row {
	result := make(database.OrderedMap, 0, len(s.fields))

	// Add GroupBy field if defined - No, we iterate over FIELDS.
	// We rely on fields order.

	for i, f := range s.fields {
		key := f.Alias
		if key == "" {
			key = f.Path
		}

		var val interface{}
		if f.Aggregate != "" {
			val = s.aggs[keyFor(i)].Result()
		} else {
			// Non-aggregated field.
			if f.Path == groupByField {
				val = groupKey
			} else {
				val = nil
			}
		}
		result = append(result, database.KeyVal{Key: key, Val: val})
	}
	return database.NewJSONRow(result)
}

// Field Aggregators

type fieldAggregator interface {
	Add(val interface{})
	Result() interface{}
}

func createAggregator(funcName string) fieldAggregator {
	switch funcName {
	case "MAX":
		return &maxAggregator{}
	case "MIN":
		return &minAggregator{}
	case "AVG":
		return &avgAggregator{}
	case "COUNT":
		return &countAggregator{}
	case "SUM":
		return &sumAggregator{}
	default:
		return &countAggregator{} // Default fallback
	}
}

// MAX
type maxAggregator struct {
	val interface{}
	set bool
}

func (a *maxAggregator) Add(v interface{}) {
	if v == nil {
		return
	}
	if slice, ok := v.([]interface{}); ok {
		for _, item := range slice {
			a.Add(item)
		}
		return
	}
	if !a.set {
		a.val = v
		a.set = true
		return
	}
	if compareGreater(v, a.val) {
		a.val = v
	}
}

func (a *maxAggregator) Result() interface{} {
	return a.val
}

// MIN
type minAggregator struct {
	val interface{}
	set bool
}

func (a *minAggregator) Add(v interface{}) {
	if v == nil {
		return
	}
	if slice, ok := v.([]interface{}); ok {
		for _, item := range slice {
			a.Add(item)
		}
		return
	}
	if !a.set {
		a.val = v
		a.set = true
		return
	}
	if compareLess(v, a.val) {
		a.val = v
	}
}

func (a *minAggregator) Result() interface{} {
	return a.val
}

// AVG
type avgAggregator struct {
	sum   float64
	count int
}

func (a *avgAggregator) Add(v interface{}) {
	if v == nil {
		return
	}
	if slice, ok := v.([]interface{}); ok {
		for _, item := range slice {
			a.Add(item)
		}
		return
	}
	f, ok := toFloat64(v)
	if ok {
		a.sum += f
		a.count++
	}
}

func (a *avgAggregator) Result() interface{} {
	if a.count == 0 {
		return 0.0
	}
	return a.sum / float64(a.count)
}

// COUNT
type countAggregator struct {
	count int
}

func (a *countAggregator) Add(v interface{}) {
	if v != nil {
		if slice, ok := v.([]interface{}); ok {
			a.count += len(slice)
		} else {
			a.count++
		}
	}
}

func (a *countAggregator) Result() interface{} {
	return a.count
}

// SUM
type sumAggregator struct {
	sum float64
}

func (a *sumAggregator) Add(v interface{}) {
	if v == nil {
		return
	}
	if slice, ok := v.([]interface{}); ok {
		for _, item := range slice {
			a.Add(item)
		}
		return
	}
	f, ok := toFloat64(v)
	if ok {
		a.sum += f
	}
}

func (a *sumAggregator) Result() interface{} {
	return a.sum
}

// Comparison Helpers (Duplicated from query/query.go or should be exported?)
// For now, simple local implementation to avoid circular deps if query imports parser/engine.
// Wait, engine imports query. So engine can use query.Compare...
// But query.go doesn't export comparison functions.
// I'll re-implement basic float comparison here.

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
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func compareGreater(a, b interface{}) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af > bf
	}
	// String comparison
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return as > bs
}

func compareLess(a, b interface{}) bool {
	af, aok := toFloat64(a)
	bf, bok := toFloat64(b)
	if aok && bok {
		return af < bf
	}
	// String comparison
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return as < bs
}
