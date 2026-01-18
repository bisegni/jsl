package engine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bisegni/jsl/pkg/database"
)

type MockRow struct {
	data map[string]interface{}
}

func (r *MockRow) Get(field string) (interface{}, error) {
	if val, ok := r.data[field]; ok {
		return val, nil
	}
	return nil, nil // proper error handling skipped for mock
}

func (r *MockRow) Primitive() interface{} {
	return r.data
}

type MockTable struct {
	rows []database.Row
}

func (t *MockTable) Iterate() (database.RowIterator, error) {
	return &MockIterator{rows: t.rows, index: -1}, nil
}

type MockIterator struct {
	rows  []database.Row
	index int
}

func (it *MockIterator) Next() bool {
	it.index++
	return it.index < len(it.rows)
}

func (it *MockIterator) Row() database.Row {
	return it.rows[it.index]
}

func (it *MockIterator) Error() error { return nil }
func (it *MockIterator) Close() error { return nil }

func TestExecutorFilter(t *testing.T) {
	rows := []database.Row{
		&MockRow{data: map[string]interface{}{"age": float64(20), "name": "Alice"}},
		&MockRow{data: map[string]interface{}{"age": float64(30), "name": "Bob"}},
	}
	table := &MockTable{rows: rows}

	executor := NewExecutor()
	// Query: SELECT * WHERE age > 25
	q := &Query{
		Fields:    []Field{},
		Condition: "age > 25",
	}

	var buf bytes.Buffer
	err := executor.Execute(q, table, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "Alice") {
		t.Errorf("Expected Alice to be filtered out, got: %s", out)
	}
	if !strings.Contains(out, "Bob") {
		t.Errorf("Expected Bob to be present, got: %s", out)
	}
}

func TestExecutorNestedFilter(t *testing.T) {
	// Simulate nested structure
	// .data.value > 10
	rows := []database.Row{
		&MockRow{data: map[string]interface{}{
			"data": map[string]interface{}{"value": float64(5)},
		}},
		&MockRow{data: map[string]interface{}{
			"data": map[string]interface{}{"value": float64(15)},
		}},
	}
	table := &MockTable{rows: rows}

	executor := NewExecutor()
	q := &Query{
		Fields:    []Field{},
		Condition: "data.value > 10",
	}

	var buf bytes.Buffer
	err := executor.Execute(q, table, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	out := buf.String()
	// First row (value 5) should be gone, second (value 15) present
	if strings.Contains(out, ": 5") {
		t.Errorf("Expected value 5 to be filtered out, got: %s", out)
	}
	if !strings.Contains(out, ": 15") {
		t.Errorf("Expected value 15 to be present, got: %s", out)
	}
}

func TestExecutorLeadingDot(t *testing.T) {
	rows := []database.Row{
		&MockRow{data: map[string]interface{}{"val": float64(30)}},
	}
	table := &MockTable{rows: rows}

	executor := NewExecutor()
	// Query with leading dot in WHERE
	q := &Query{
		Fields:    []Field{},
		Condition: ".val > 20",
	}

	var buf bytes.Buffer
	err := executor.Execute(q, table, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "30") {
		t.Errorf("Expected 30 to be present with leading dot filter, got: %s", out)
	}
}

func TestExecutorProjectionOutput(t *testing.T) {
	row := database.NewJSONRow(map[string]interface{}{
		"sensors": []interface{}{
			map[string]interface{}{"name": "A", "type": "temp"},
			map[string]interface{}{"name": "B", "type": "humidity"},
		},
	})

	table := &MockTable{rows: []database.Row{row}}

	exec := NewExecutor()

	// Query: SELECT sensors.*.type='temp' AS temp_sensors
	q := &Query{
		Fields: []Field{{Path: "sensors.*.type='temp'", Alias: "temp_sensors"}},
	}

	var buf bytes.Buffer
	err := exec.Execute(q, table, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := buf.String()
	t.Logf("Output: %s", output)

	// Check if key is "temp_sensors"
	if !strings.Contains(output, "temp_sensors") {
		t.Errorf("Expected output to contain alias 'temp_sensors', got: %s", output)
	}
}

func TestExecutorRefinedProjection(t *testing.T) {
	row := database.NewJSONRow(map[string]interface{}{
		"sensors": []interface{}{
			map[string]interface{}{"name": "A", "type": "temp"},
			map[string]interface{}{"name": "B", "type": "humidity"},
		},
	})
	table := &MockTable{rows: []database.Row{row}}

	exec := NewExecutor()

	// Query: SELECT sensors.$.name WHERE sensors.*.type='temp'
	q := &Query{
		Fields:    []Field{{Path: "sensors.$.name", Alias: "matched_name"}},
		Condition: "sensors.*.type='temp'",
	}

	var buf bytes.Buffer
	err := exec.Execute(q, table, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := buf.String()
	t.Logf("Output: %s", output)

	// Expect only name "A" (matched temp), not "B" (humidity)
	if !strings.Contains(output, "A") {
		t.Error("Expected output to contain 'A'")
	}
	if strings.Contains(output, "B") {
		t.Error("Expected output to NOT contain 'B'")
	}
}
