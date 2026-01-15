package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [file|-]",
	Short: "Show statistics about JSON/JSONL file",
	Long: `Display statistics about a JSON or JSONL file including record count,
field types, and structure information.

Supports:
  - File paths: jsl stats data.json
  - Stdin: cat data.json | jsl stats

Examples:
  jsl stats data.json
  jsl stats data.jsonl
  cat data.json | jsl stats`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
	filename := "-"
	if len(args) > 0 {
		filename = args[0]
	}

	p, err := parser.NewParser(filename)
	if err != nil {
		return err
	}
	defer p.Close()

	records, err := p.ReadAll()
	if err != nil {
		return err
	}

	// Gather statistics
	stats := gatherStats(records)

	// Print statistics
	if filename == "-" {
		fmt.Printf("File: <stdin>\n")
	} else {
		fmt.Printf("File: %s\n", filename)
	}
	fmt.Printf("Format: %s\n", getFormat(p.IsJSONL()))
	fmt.Printf("Total records: %d\n", stats["total_records"])
	
	if fields, ok := stats["fields"].(map[string]map[string]int); ok {
		fmt.Printf("\nFields:\n")
		for field, types := range fields {
			fmt.Printf("  %s:\n", field)
			for typ, count := range types {
				fmt.Printf("    %s: %d (%.1f%%)\n", typ, count, float64(count)/float64(stats["total_records"].(int))*100)
			}
		}
	}

	return nil
}

func getFormat(isJSONL bool) string {
	if isJSONL {
		return "JSONL"
	}
	return "JSON"
}

func gatherStats(records []parser.Record) map[string]interface{} {
	stats := map[string]interface{}{
		"total_records": len(records),
		"fields":        make(map[string]map[string]int),
	}

	fields := make(map[string]map[string]int)

	for _, record := range records {
		for key, value := range record {
			if _, exists := fields[key]; !exists {
				fields[key] = make(map[string]int)
			}
			
			typeName := getTypeName(value)
			fields[key][typeName]++
		}
	}

	stats["fields"] = fields
	return stats
}

func getTypeName(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch v.(type) {
	case bool:
		return "boolean"
	case float64, int, int64, float32:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		// Use JSON marshaling to determine type
		data, _ := json.Marshal(v)
		var test interface{}
		json.Unmarshal(data, &test)
		switch test.(type) {
		case []interface{}:
			return "array"
		case map[string]interface{}:
			return "object"
		default:
			return "unknown"
		}
	}
}
