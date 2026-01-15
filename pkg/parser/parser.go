package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Record represents a single JSON object
type Record map[string]interface{}

// Parser handles reading JSON and JSONL files
type Parser struct {
	file   *os.File
	isJSONL bool
}

// NewParser creates a new parser for the given file
// Special cases:
// - Empty string or "-" reads from stdin
// - Strings starting with '{' or '[' are treated as inline JSON
func NewParser(filename string) (*Parser, error) {
	var file *os.File
	var err error
	var isJSONL bool

	// Handle inline JSON (starts with { or [)
	if len(filename) > 0 && (filename[0] == '{' || filename[0] == '[') {
		// Create a temporary file to store inline JSON
		tmpFile, err := os.CreateTemp("", "jsl-inline-*.json")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		if _, err := tmpFile.WriteString(filename); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to write inline JSON: %w", err)
		}
		// Seek back to the beginning
		if _, err := tmpFile.Seek(0, 0); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, fmt.Errorf("failed to seek: %w", err)
		}
		file = tmpFile
		isJSONL = false
	} else if filename == "" || filename == "-" {
		// Read from stdin
		file = os.Stdin
		isJSONL = false // Auto-detect by trying JSON first
	} else {
		// Regular file
		file, err = os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		// Try to detect if it's JSONL by checking file extension
		isJSONL = len(filename) >= 6 && filename[len(filename)-6:] == ".jsonl"
	}

	return &Parser{
		file:   file,
		isJSONL: isJSONL,
	}, nil
}

// Close closes the underlying file
func (p *Parser) Close() error {
	return p.file.Close()
}

// IsJSONL returns whether the parser is treating the file as JSONL
func (p *Parser) IsJSONL() bool {
	return p.isJSONL
}

// ReadAll reads all records from the file
func (p *Parser) ReadAll() ([]Record, error) {
	if p.isJSONL {
		return p.readJSONL()
	}
	return p.readJSON()
}

// readJSON reads a single JSON file
func (p *Parser) readJSON() ([]Record, error) {
	decoder := json.NewDecoder(p.file)
	
	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to array of records
	switch v := data.(type) {
	case map[string]interface{}:
		// Single object
		return []Record{v}, nil
	case []interface{}:
		// Array of objects
		records := make([]Record, 0, len(v))
		for i, item := range v {
			if obj, ok := item.(map[string]interface{}); ok {
				records = append(records, obj)
			} else {
				return nil, fmt.Errorf("array element %d is not an object", i)
			}
		}
		return records, nil
	default:
		return nil, fmt.Errorf("unexpected JSON type: %T", v)
	}
}

// readJSONL reads a JSONL (JSON Lines) file
func (p *Parser) readJSONL() ([]Record, error) {
	var records []Record
	scanner := bufio.NewScanner(p.file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue // Skip empty lines
		}

		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return records, nil
}

// ForEachRecord processes each record with the given function
func (p *Parser) ForEachRecord(fn func(Record) error) error {
	if p.isJSONL {
		return p.forEachJSONL(fn)
	}
	return p.forEachJSON(fn)
}

func (p *Parser) forEachJSON(fn func(Record) error) error {
	records, err := p.readJSON()
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := fn(record); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) forEachJSONL(fn func(Record) error) error {
	scanner := bufio.NewScanner(p.file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		if err := fn(record); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// WriteJSON writes records as a JSON array
func WriteJSON(w io.Writer, records []Record, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(records)
}

// WriteJSONL writes records as JSON Lines
func WriteJSONL(w io.Writer, records []Record, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}
