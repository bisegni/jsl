package cmd

import (
	"encoding/json"
	"os"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
	"github.com/spf13/cobra"
)

var (
	queryPath   string
	queryPretty bool
)

var queryCmd = &cobra.Command{
	Use:   "query [file]",
	Short: "Query JSON/JSONL file with path expression",
	Long: `Query a JSON or JSONL file using a dot-separated path expression.
Examples:
  jsl query data.json --path .user.name
  jsl query data.jsonl --path .items.*.price
  jsl query data.json --path .metadata`,
	Args: cobra.ExactArgs(1),
	RunE: runQuery,
}

func init() {
	queryCmd.Flags().StringVarP(&queryPath, "path", "p", ".", "Path to extract (e.g., .user.name)")
	queryCmd.Flags().BoolVar(&queryPretty, "pretty", true, "Pretty print output")
}

func runQuery(cmd *cobra.Command, args []string) error {
	filename := args[0]

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

	if len(results) == 1 {
		return encoder.Encode(results[0])
	}
	return encoder.Encode(results)
}
