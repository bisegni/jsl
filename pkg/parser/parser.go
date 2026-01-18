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
	file    *os.File
	isJSONL bool
	tmpFile string // Path to temporary file, if created

	// Stateful readers
	decoder   *json.Decoder
	scanner   *bufio.Scanner
	bufReader *bufio.Reader

	startArrayChecked bool
	inArray           bool
}

// NewParser creates a new parser for the given file
// Special cases:
// - Empty string or "-" reads from stdin
// - Strings starting with '{' or '[' are treated as inline JSON
func NewParser(filename string) (*Parser, error) {
	var file *os.File
	var err error
	var isJSONL bool
	var tmpFile string

	// Handle inline JSON (starts with { or [)
	if len(filename) > 0 && (filename[0] == '{' || filename[0] == '[') {
		// Create a temporary file to store inline JSON
		tmpFileHandle, err := os.CreateTemp("", "jsl-inline-*.json")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpFile = tmpFileHandle.Name()
		if _, err := tmpFileHandle.WriteString(filename); err != nil {
			tmpFileHandle.Close()
			os.Remove(tmpFile)
			return nil, fmt.Errorf("failed to write inline JSON: %w", err)
		}
		// Seek back to the beginning
		if _, err := tmpFileHandle.Seek(0, 0); err != nil {
			tmpFileHandle.Close()
			os.Remove(tmpFile)
			return nil, fmt.Errorf("failed to seek: %w", err)
		}
		file = tmpFileHandle
		isJSONL = false
	} else if filename == "" || filename == "-" {
		// Read from stdin
		file = os.Stdin
		isJSONL = false // Default to false, will try auto-detect if needed? No, logic below.
	} else {
		// Regular file
		file, err = os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		// Try to detect if it's JSONL by checking file extension
		isJSONL = len(filename) >= 6 && filename[len(filename)-6:] == ".jsonl"
	}

	p := &Parser{
		file:    file,
		isJSONL: isJSONL,
		tmpFile: tmpFile,
	}

	p.initReader()
	return p, nil
}

func (p *Parser) initReader() {
	if p.isJSONL {
		p.scanner = bufio.NewScanner(p.file)
	} else {
		// Use bufio.Reader to allow peeking
		p.bufReader = bufio.NewReader(p.file)
		p.decoder = json.NewDecoder(p.bufReader)
	}
}

// Close closes the underlying file and cleans up any temporary files
func (p *Parser) Close() error {
	err := p.file.Close()
	// Clean up temporary file if it exists
	if p.tmpFile != "" {
		os.Remove(p.tmpFile)
	}
	return err
}

// IsJSONL returns whether the parser is treating the file as JSONL
func (p *Parser) IsJSONL() bool {
	return p.isJSONL
}

// Read reads the next record from the file.
func (p *Parser) Read() (Record, error) {
	if p.isJSONL {
		if !p.scanner.Scan() {
			if err := p.scanner.Err(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		line := p.scanner.Text()
		if len(line) == 0 {
			return p.Read()
		}
		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("failed to parse JSONL record: %w", err)
		}
		return record, nil
	}

	// Standard JSON Streaming Logic
	if !p.startArrayChecked {
		// Peek first non-whitespace byte
		for {
			b, err := p.bufReader.Peek(1)
			if err != nil {
				if err == io.EOF {
					return nil, io.EOF
				}
				return nil, err
			}
			c := b[0]
			if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
				p.bufReader.ReadByte() // consume whitespace
				continue
			}
			if c == '[' {
				p.inArray = true
				p.bufReader.ReadByte() // consume '['
			}
			p.startArrayChecked = true
			break
		}
	}

	if p.inArray {
		if !p.decoder.More() {
			// Consume closing ']'
			t, err := p.decoder.Token()
			if err != nil {
				return nil, err
			}
			if delim, ok := t.(json.Delim); ok && delim == ']' {
				p.inArray = false
				return nil, io.EOF
			}
			return nil, fmt.Errorf("expected array end, got %v", t)
		}
	} else {
		// Stream of objects or single object
		// Check EOF via peek, because decoder.More() might rely on array delimiters?
		// No, More() is for arrays.
		// Just try Decode.
		// But check EOF first because Decode might return EOF after reading whitespace.
	}

	// Decode next item
	var record Record
	if err := p.decoder.Decode(&record); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to decode JSON record: %w", err)
	}
	return record, nil
}

// ReadAll reads all records from the file
// This maintains backward compatibility by using the robust logic
func (p *Parser) ReadAll() ([]Record, error) {
	// Re-open/seek if we read partially?
	// For safety, let's just delegate to existing logic but separate impl?
	// Or try to use the reader.
	// Given the database refactor, let's keep the existing implementation structure for ReadAll
	// but make sure it creates a fresh independent reader or resets.
	// But we can't easily reset stdin.

	if p.isJSONL {
		return p.readJSONL()
	}
	return p.readJSON()
}

// readJSON reads a single JSON file
func (p *Parser) readJSON() ([]Record, error) {
	// Reset decoder?
	p.file.Seek(0, 0)
	decoder := json.NewDecoder(p.file)

	var allRecords []Record

	for {
		var data interface{}
		if err := decoder.Decode(&data); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Convert to array of records
		switch v := data.(type) {
		case map[string]interface{}:
			// Single object
			allRecords = append(allRecords, v)
		case []interface{}:
			// Array of objects
			for i, item := range v {
				if obj, ok := item.(map[string]interface{}); ok {
					allRecords = append(allRecords, obj)
				} else {
					return nil, fmt.Errorf("array element %d is not an object", i)
				}
			}
		default:
			return nil, fmt.Errorf("unexpected JSON type: %T", v)
		}
	}
	return allRecords, nil
}

// readJSONL reads a JSONL (JSON Lines) file
func (p *Parser) readJSONL() ([]Record, error) {
	// Reset scanner?
	p.file.Seek(0, 0)
	scanner := bufio.NewScanner(p.file)

	var records []Record
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("failed to parse JSONL record: %w", err)
		}
		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSONL file: %w", err)
	}

	return records, nil
}

// ForEachRecord processes each record with the given function
func (p *Parser) ForEachRecord(fn func(Record) error) error {
	// For compatibility, use ReadAll logic
	records, err := p.ReadAll()
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
