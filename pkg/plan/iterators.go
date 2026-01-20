package plan

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
)

// --- Filter Iterator ---

type filterIterator struct {
	source     database.RowIterator
	expression query.Expression
}

func (it *filterIterator) Next() bool {
	for it.source.Next() {
		// Convert Row back to Record for Match
		primitive := it.source.Row().Primitive()

		var record map[string]interface{}
		switch v := primitive.(type) {
		case parser.Record:
			record = v
		case map[string]interface{}:
			record = v
		case database.OrderedMap:
			record = v.ToMap()
		default:
			continue
		}

		if it.expression.Evaluate(record) {
			return true
		}
	}
	return false
}

func (it *filterIterator) Row() database.Row {
	return it.source.Row()
}

func (it *filterIterator) Error() error {
	return it.source.Error()
}

func (it *filterIterator) Close() error {
	return it.source.Close()
}

// --- Project Iterator ---

type projectIterator struct {
	source      database.RowIterator
	fields      []query.Field
	currentRow  database.Row
	pendingRows []database.Row
}

func (it *projectIterator) Next() bool {
	// 1. Check if we have pending rows from significant unwinding
	if len(it.pendingRows) > 0 {
		it.currentRow = it.pendingRows[0]
		it.pendingRows = it.pendingRows[1:]
		return true
	}

	// 2. Fetch corresponding next row from source
	if it.source.Next() {
		srcRow := it.source.Row()

		type fieldVal struct {
			key      string
			val      interface{}
			isArray  bool
			arrayVal []interface{}
		}

		fVals := make([]fieldVal, len(it.fields))

		allArraysLength := -1
		consistentArrays := true
		hasArrays := false

		for i, f := range it.fields {
			key := f.Alias
			if key == "" {
				key = f.Path
			}

			val, err := srcRow.Get(f.Path)
			if err != nil {
				val = nil
			}

			fv := fieldVal{key: key, val: val}

			if sliceVal, ok := val.([]interface{}); ok {
				fv.isArray = true
				fv.arrayVal = sliceVal
				hasArrays = true

				if allArraysLength == -1 {
					allArraysLength = len(sliceVal)
				} else if allArraysLength != len(sliceVal) {
					consistentArrays = false
				}
			}
			fVals[i] = fv
		}

		// 3. Unwind Logic
		if hasArrays && consistentArrays && allArraysLength > 0 {
			// Generate N rows
			for i := 0; i < allArraysLength; i++ {
				// Build OrderedMap
				newRow := make(database.OrderedMap, len(it.fields))
				for j, fv := range fVals {
					var v interface{}
					if fv.isArray {
						v = fv.arrayVal[i]
					} else {
						v = fv.val
					}
					newRow[j] = database.KeyVal{Key: fv.key, Val: v}
				}
				it.pendingRows = append(it.pendingRows, database.NewJSONRow(newRow))
			}

			it.currentRow = it.pendingRows[0]
			it.pendingRows = it.pendingRows[1:]
			return true
		}

		// 4. Fallback: Return as is
		newRow := make(database.OrderedMap, len(it.fields))
		for i, fv := range fVals {
			newRow[i] = database.KeyVal{Key: fv.key, Val: fv.val}
		}
		it.currentRow = database.NewJSONRow(newRow)
		return true
	}
	return false
}

func (it *projectIterator) Row() database.Row {
	return it.currentRow
}

func (it *projectIterator) Error() error {
	return it.source.Error()
}

func (it *projectIterator) Close() error {
	return it.source.Close()
}

// --- Aggregate Iterator ---

type aggregateIterator struct {
	input        Node
	groupByField string
	fields       []query.Field

	results []database.Row
	index   int
}

func (it *aggregateIterator) Next() bool {
	// Initialize on first call
	if it.results == nil {
		if err := it.init(); err != nil {
			return false
		}
	}
	it.index++
	return it.index < len(it.results)
}

func (it *aggregateIterator) Row() database.Row {
	if it.index >= 0 && it.index < len(it.results) {
		return it.results[it.index]
	}
	return nil
}

func (it *aggregateIterator) Error() error {
	return nil // Initialization error handled in Next() ?? TODO: persist error
}

func (it *aggregateIterator) Close() error {
	return nil
}

func (it *aggregateIterator) init() error {
	sourceIter, err := it.input.Execute()
	if err != nil {
		return err
	}
	defer sourceIter.Close()

	groups := make(map[string]*groupState)
	var groupKeys []string
	hasData := false

	extract := func(row database.Row, path string) (interface{}, error) {
		return row.Get(path)
	}

	for sourceIter.Next() {
		hasData = true
		row := sourceIter.Row()

		var groupKey string
		if it.groupByField != "" {
			val, err := extract(row, it.groupByField)
			if err == nil {
				groupKey = fmt.Sprintf("%v", val)
			} else {
				groupKey = "null"
			}
		} else {
			groupKey = ""
		}

		state, exists := groups[groupKey]
		if !exists {
			state = newGroupState(it.fields)
			groups[groupKey] = state
			groupKeys = append(groupKeys, groupKey)
		}

		state.update(row, extract)
	}

	if err := sourceIter.Error(); err != nil {
		return err
	}

	// Build results
	it.results = []database.Row{}
	it.index = -1

	// Handle empty input with global aggregation
	if !hasData && it.groupByField == "" && len(it.fields) > 0 {
		hasAgg := false
		for _, f := range it.fields {
			if f.Aggregate != "" {
				hasAgg = true
				break
			}
		}
		if hasAgg {
			state := newGroupState(it.fields)
			it.results = append(it.results, state.finalize("", ""))
			return nil
		}
	}

	sort.Strings(groupKeys)

	for _, key := range groupKeys {
		state := groups[key]
		it.results = append(it.results, state.finalize(key, it.groupByField))
	}

	return nil
}

type groupState struct {
	fields []query.Field
	aggs   map[string]fieldAggregator
}

func newGroupState(fields []query.Field) *groupState {
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
		if f.Aggregate != "" {
			val, err := extractor(row, f.Path)
			if err == nil {
				s.aggs[keyFor(i)].Add(val)
			}
		}
	}
}

func (s *groupState) finalize(groupKey string, groupByField string) database.Row {
	result := make(database.OrderedMap, len(s.fields))
	for i, f := range s.fields {
		key := f.Alias
		if key == "" {
			key = f.Path
		}
		var val interface{}
		if f.Aggregate != "" {
			val = s.aggs[keyFor(i)].Result()
		} else {
			if f.Path == groupByField {
				val = groupKey
			} else {
				val = nil
			}
		}
		result[i] = database.KeyVal{Key: key, Val: val}
	}
	return database.NewJSONRow(result)
}

// Aggregators
type fieldAggregator interface {
	Add(val interface{})
	Result() interface{}
}

func createAggregator(funcName string) fieldAggregator {
	switch strings.ToUpper(funcName) {
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
		return &countAggregator{}
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

// Helpers
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
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return as < bs
}
