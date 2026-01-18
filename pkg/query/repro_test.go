package query

import (
	"encoding/json"
	"testing"

	"github.com/bisegni/jsl/pkg/parser"
)

func TestAdvancedProjection(t *testing.T) {
	jsonStr := `{
		"timestamp": "2026-01-15T10:00:00Z", 
		"sensors": [
			{"name": "sensor_01", "type": "temp", "room": "living", "value": 22.5}, 
			{"name": "sensor_02", "type": "humidity", "room": "living", "value": 45.0}, 
			{"name": "sensor_03", "type": "temp", "room": "kitchen", "value": 23.1}
		]
	}`

	var record parser.Record
	if err := json.Unmarshal([]byte(jsonStr), &record); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Test 1: Select filtered array
	path := "sensors.*.type=temp"
	q := NewQuery(path)
	val, err := q.Extract(record)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	t.Logf("Path: %s", path)
	t.Logf("Result: %+v", val)

	// Verify result is array of length 2
	resultSlice, ok := val.([]interface{})
	if !ok {
		t.Errorf("Expected slice, got %T", val)
	} else if len(resultSlice) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resultSlice))
	}

	// Test 2: Select subfield with filter
	path2 := "sensors.*.type=temp.name"
	q2 := NewQuery(path2)
	val2, err := q2.Extract(record)
	if err != nil {
		t.Fatalf("Extract 2 failed: %v", err)
	}

	t.Logf("Path 2: %s", path2)
	t.Logf("Result 2: %+v", val2)

	resultSlice2, ok := val2.([]interface{})
	if !ok {
		t.Errorf("Expected slice, got %T", val2)
	} else {
		// Expect ["sensor_01", "sensor_03"]
		if len(resultSlice2) != 2 {
			t.Errorf("Expected 2 items, got %d", len(resultSlice2))
		}
		if resultSlice2[0] != "sensor_01" || resultSlice2[1] != "sensor_03" {
			t.Errorf("Unexpected values: %v", resultSlice2)
		}
	}

	// Test 3: Quoted value
	path3 := "sensors.*.type='temp'"
	q3 := NewQuery(path3)
	val3, err := q3.Extract(record)
	if err != nil {
		t.Fatalf("Extract 3 failed: %v", err)
	}
	t.Logf("Path 3: %s", path3)
	t.Logf("Result 3: %+v", val3)
	if slice3, ok := val3.([]interface{}); !ok || len(slice3) != 2 {
		t.Error("Failed quoted value test")
	}
}
