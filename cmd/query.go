package cmd

import (
	"encoding/json"
	"os"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
	"github.com/spf13/cobra"
)

var (
	QueryPath   string
	QueryPretty bool
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
		
		return RunQuery(filename, path, QueryPretty)
	},
}

func init() {
}

func RunQuery(filename string, queryPath string, queryPretty bool) error {
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

	// If path is "." or empty, return all records
	if queryPath == "" || queryPath == "." {
		encoder := json.NewEncoder(os.Stdout)
		if queryPretty {
			encoder.SetIndent("", "  ")
		}
		return encoder.Encode(records)
	}

	results := make([]interface{}, 0, len(records))

	for _, record := range records {
		val, err := q.Extract(record)
		if err != nil {
			continue // Skip records where path doesn't exist
		}
		results = append(results, val)
	}

	// Output results
	encoder := json.NewEncoder(os.Stdout)
	if queryPretty {
		encoder.SetIndent("", "  ")
	}

	if len(results) == 0 {
		return nil
	}

	if len(results) == 1 {
		return encoder.Encode(results[0])
	}
	return encoder.Encode(results)
}
