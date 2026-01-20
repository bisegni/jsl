package parser

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	// Create a temporary JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	content := `[{"name": "Alice", "age": 30}]`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatalf("NewParser failed: %v", err)
	}
	defer parser.Close()

	if parser.IsJSONL() {
		t.Error("Expected JSON file to not be detected as JSONL")
	}
}

func TestReadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	content := `[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}]`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0]["name"] != "Alice" {
		t.Errorf("Expected first record name to be Alice, got %v", records[0]["name"])
	}
}

func TestReadJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{"name": "Alice", "age": 30}
{"name": "Bob", "age": 25}`
	if err := os.WriteFile(jsonlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonlFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0]["name"] != "Alice" {
		t.Errorf("Expected first record name to be Alice, got %v", records[0]["name"])
	}

	if !parser.IsJSONL() {
		t.Error("Expected JSONL file to be detected as JSONL")
	}
}

func TestReadJSONSingleObject(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	content := `{"name": "Alice", "age": 30}`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(records))
	}

	if records[0]["name"] != "Alice" {
		t.Errorf("Expected record name to be Alice, got %v", records[0]["name"])
	}
}

func TestReadJSONNested(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "nested.json")

	content := `[
		{"name": "Alice", "info": {"city": "New York", "hobbies": ["reading", "cycling"]}},
		{"name": "Bob", "info": {"city": "London", "hobbies": ["drawing"]}}
	]`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if info, ok := records[0]["info"].(map[string]interface{}); ok {
		if info["city"] != "New York" {
			t.Errorf("Expected city New York, got %v", info["city"])
		}
	} else {
		t.Errorf("Expected info to be a map, got %T", records[0]["info"])
	}
}

func TestReadJSONConcatenated(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "concat.json")

	content := `{"name": "Alice"}{"name": "Bob"}`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestReadJSONMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "malformed.json")

	content := `[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	_, err = parser.ReadAll()
	if err == nil {
		t.Error("Expected error for malformed JSON, got nil")
	}
}

func TestReadJSONLMalformed(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlFile := filepath.Join(tmpDir, "malformed.jsonl")

	content := `{"name": "Alice"}
{"name": "Bob", "age": 25
{"name": "Charlie"}`
	if err := os.WriteFile(jsonlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonlFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	_, err = parser.ReadAll()
	if err == nil {
		t.Error("Expected error for malformed JSONL line, got nil")
	}
}

func TestReadJSONLEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlFile := filepath.Join(tmpDir, "empty_lines.jsonl")

	content := `{"name": "Alice"}

{"name": "Bob"}
`
	if err := os.WriteFile(jsonlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonlFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}

func TestInlineJSON(t *testing.T) {
	content := `[{"name": "Alice"}, {"name": "Bob"}]`
	parser, err := NewParser(content)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if parser.IsJSONL() {
		t.Error("Expected inline JSON to not be detected as JSONL")
	}
}

func TestEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "empty.json")

	if err := os.WriteFile(jsonFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(jsonFile)
	if err != nil {
		t.Fatal(err)
	}
	defer parser.Close()

	records, err := parser.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed for empty file: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records for empty file, got %d", len(records))
	}
}

func TestReadStreaming(t *testing.T) {
	t.Run("JSONL", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlFile := filepath.Join(tmpDir, "stream.jsonl")
		content := `{"id": 1}
{"id": 2}
{"id": 3}`
		os.WriteFile(jsonlFile, []byte(content), 0644)

		parser, _ := NewParser(jsonlFile)
		defer parser.Close()

		var count int
		for {
			rec, err := parser.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("Read failed: %v", err)
			}
			count++
			if int(rec["id"].(float64)) != count {
				t.Errorf("Expected id %d, got %v", count, rec["id"])
			}
		}
		if count != 3 {
			t.Errorf("Expected 3 records, got %d", count)
		}
	})

	t.Run("JSONArray", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonFile := filepath.Join(tmpDir, "stream.json")
		content := `[{"id": 1}, {"id": 2}, {"id": 3}]`
		os.WriteFile(jsonFile, []byte(content), 0644)

		parser, _ := NewParser(jsonFile)
		defer parser.Close()

		var count int
		for {
			rec, err := parser.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("Read failed: %v", err)
			}
			count++
			if int(rec["id"].(float64)) != count {
				t.Errorf("Expected id %d, got %v", count, rec["id"])
			}
		}
		if count != 3 {
			t.Errorf("Expected 3 records, got %d", count)
		}
	})
}
