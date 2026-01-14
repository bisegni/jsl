package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/bisegni/jsl/pkg/query"
	"github.com/spf13/cobra"
)

var (
	filterField    string
	filterOperator string
	filterValue    string
	filterPretty   bool
	filterFormat   string
)

var filterCmd = &cobra.Command{
	Use:   "filter [file]",
	Short: "Filter JSON/JSONL records based on conditions",
	Long: `Filter records from a JSON or JSONL file based on field conditions.
Examples:
  jsl filter data.json --field .age --op ">" --value 18
  jsl filter data.jsonl --field .status --op "=" --value active
  jsl filter data.json --field .name --op contains --value john`,
	Args: cobra.ExactArgs(1),
	RunE: runFilter,
}

func init() {
	filterCmd.Flags().StringVarP(&filterField, "field", "f", "", "Field path to filter on")
	filterCmd.Flags().StringVarP(&filterOperator, "op", "o", "=", "Operator (=, !=, >, >=, <, <=, contains)")
	filterCmd.Flags().StringVarP(&filterValue, "value", "v", "", "Value to compare against")
	filterCmd.Flags().BoolVar(&filterPretty, "pretty", true, "Pretty print output")
	filterCmd.Flags().StringVar(&filterFormat, "format", "json", "Output format (json or jsonl)")
	filterCmd.MarkFlagRequired("field")
	filterCmd.MarkFlagRequired("value")
}

func runFilter(cmd *cobra.Command, args []string) error {
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

	// Parse filter value
	var filterVal interface{}
	filterVal = filterValue
	
	// Try to parse as number
	if val, err := parseNumber(filterValue); err == nil {
		filterVal = val
	}

	f := query.NewFilter(filterField, filterOperator, filterVal)
	var filtered []parser.Record

	for _, record := range records {
		if f.Match(record) {
			filtered = append(filtered, record)
		}
	}

	// Output filtered records
	if strings.ToLower(filterFormat) == "jsonl" {
		return parser.WriteJSONL(os.Stdout, filtered, filterPretty)
	}
	return parser.WriteJSON(os.Stdout, filtered, filterPretty)
}

func parseNumber(s string) (interface{}, error) {
	var val interface{}
	if err := json.Unmarshal([]byte(s), &val); err != nil {
		return nil, err
	}
	switch val.(type) {
	case float64, int, int64:
		return val, nil
	default:
		return nil, fmt.Errorf("not a number")
	}
}
