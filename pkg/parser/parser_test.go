package parser

import (
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

func TestReadMultiLineJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	jsonlFile := filepath.Join(tmpDir, "test.jsonl")

	content := `{
  "name": "Alice",
  "age": 30
}
{
  "name": "Bob",
  "age": 25
}`
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

	if records[1]["name"] != "Bob" {
		t.Errorf("Expected second record name to be Bob, got %v", records[1]["name"])
	}
}
