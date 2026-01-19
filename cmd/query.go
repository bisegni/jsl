package cmd

import (
	"encoding/json"
	"os"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [file|JSON|-] [path]",
	Short: "Query JSON/JSONL file with path expression",
	Long: `Query a JSON or JSONL file using a dot-separated path expression.

Supports:
  - File paths: jsl query data.json .user.name
  - Stdin: cat data.json | jsl query - .user.name (or omit filename)
  - Inline JSON: jsl query '{"user":{"name":"Alice"}}' .user.name

Examples:
  jsl query data.json .user.name
  jsl query data.jsonl .items.*.price
  cat data.json | jsl query - .metadata
  echo '{"name":"Alice"}' | jsl query .name
  jsl query '{"user":{"name":"Alice"}}' .user.name`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle different argument patterns
		var filename, path string

		if len(args) == 0 {
			// No args, read from stdin
			filename = "-"
			path = QueryPath
		} else if len(args) == 1 {
			// One arg: could be filename or path
			arg := args[0]
			// If it looks like a path (starts with .) or stdin marker, treat as path
			if arg == "-" || (len(arg) > 0 && arg[0] == '.') {
				filename = "-"
				path = arg
			} else {
				// Otherwise it's a filename
				filename = arg
				path = QueryPath
			}
		} else {
			// Two args: filename and path
			filename = args[0]
			path = args[1]
		}

		return RunQuery(filename, path, QueryPretty, QueryExtract, QuerySelect)
	},
}

func init() {
}

func RunQuery(filename string, queryPath string, queryPretty bool, queryExtract bool, selectFields []string) error {
	p, err := parser.NewParser(filename)
	if err != nil {
		return err
	}
	defer p.Close()

	records, err := p.ReadAll()
	if err != nil {
		return err
	}

	q := query.NewQuery(queryPath)

	// If path is "." or empty, apply selection to all records
	if queryPath == "" || queryPath == "." {
		encoder := json.NewEncoder(os.Stdout)
		if queryPretty {
			encoder.SetIndent("", "  ")
		} else {
			encoder.SetIndent("", "")
		}

		for _, record := range records {
			var output interface{}
			if len(selectFields) > 0 {
				output = applySelection(record, selectFields)
			} else {
				output = record
			}
			if err := encoder.Encode(output); err != nil {
				return err
			}
		}
		return nil
	}

	// Output results
	encoder := json.NewEncoder(os.Stdout)
	if queryPretty {
		encoder.SetIndent("", "  ")
	} else {
		encoder.SetIndent("", "")
	}

	for _, record := range records {
		val, err := q.Extract(record)
		if err != nil {
			continue // Skip records where path doesn't exist
		}

		var resultsToPrint []interface{}

		if queryExtract {
			switch v := val.(type) {
			case map[string]interface{}:
				for k, subVal := range v {
					if len(selectFields) > 0 {
						item := applySelection(subVal, selectFields)
						resultsToPrint = append(resultsToPrint, item)
					} else {
						resultsToPrint = append(resultsToPrint, map[string]interface{}{k: subVal})
					}
				}
			case []interface{}:
				for _, item := range v {
					if len(selectFields) > 0 {
						item = applySelection(item, selectFields)
					}
					resultsToPrint = append(resultsToPrint, item)
				}
			default:
				if len(selectFields) > 0 {
					val = applySelection(val, selectFields)
				}
				resultsToPrint = append(resultsToPrint, val)
			}
		} else {
			if len(selectFields) > 0 {
				val = applySelection(val, selectFields)
			}
			resultsToPrint = append(resultsToPrint, val)
		}

		for _, res := range resultsToPrint {
			if err := encoder.Encode(res); err != nil {
				return err
			}
		}
	}

	return nil
}

func applySelection(val interface{}, fields []string) interface{} {
	switch v := val.(type) {
	case parser.Record:
		newMap := make(parser.Record)
		for _, f := range fields {
			if val, ok := v[f]; ok {
				newMap[f] = val
			}
		}
		return newMap
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for _, f := range fields {
			if val, ok := v[f]; ok {
				newMap[f] = val
			}
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, item := range v {
			newSlice[i] = applySelection(item, fields)
		}
		return newSlice
	default:
		return val
	}
}
