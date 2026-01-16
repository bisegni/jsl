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
	filterField      string
	filterOperator   string
	filterValue      string
	filterPretty     bool
	filterFormat     string
	filterExpression string
)

var filterCmd = &cobra.Command{
	Use:   "filter [file|-] [expression]",
	Short: "Filter JSON/JSONL records based on conditions",
	Long: `Filter records from a JSON or JSONL file based on field conditions.

Supports two syntax styles:
1. Expression style (recommended): jsl filter data.json age>28
2. Flag style (verbose): jsl filter data.json --field age --op ">" --value 28

Expression operators: =, !=, >, >=, <, <=, ~= (contains)

Examples:
  # Expression style (concise)
  jsl filter data.json age>28
  jsl filter data.jsonl status=active
  jsl filter data.json name~=john
  cat data.json | jsl filter - age>=30
  cat data.json | jsl filter age>=30
  
  # Flag style (verbose)
  jsl filter data.json --field age --op ">" --value 28
  jsl filter data.jsonl --field status --op "=" --value active`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runFilter,
}

func init() {
	filterCmd.Flags().StringVarP(&filterField, "field", "f", "", "Field path to filter on")
	filterCmd.Flags().StringVarP(&filterOperator, "op", "o", "=", "Operator (=, !=, >, >=, <, <=, contains)")
	filterCmd.Flags().StringVarP(&filterValue, "value", "v", "", "Value to compare against")
	filterCmd.Flags().BoolVar(&filterPretty, "pretty", true, "Pretty print output")
	filterCmd.Flags().StringVar(&filterFormat, "format", "json", "Output format (json or jsonl)")
}

// IsFilterExpression checks if a string looks like a filter expression (contains an operator)
// and does NOT start with a dot (which signifies a path query)
func IsFilterExpression(expr string) bool {
	return query.IsFilterExpression(expr)
}

func RunFilter(filename string, field, operator, value string, pretty bool, extract bool, selectFields []string, format string) error {
	// Validate we have all required fields
	if field == "" || value == "" {
		return fmt.Errorf("field and value are required")
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

	// Parse filter value
	var filterVal interface{}
	filterVal = value

	// Try to parse as number
	if val, err := parseNumber(value); err == nil {
		filterVal = val
	}

	f := query.NewFilter(field, operator, filterVal)
	var filtered []parser.Record

	for _, record := range records {
		if f.Match(record) {
			if len(selectFields) > 0 {
				pruned := make(parser.Record)
				for _, fld := range selectFields {
					if val, ok := record[fld]; ok {
						pruned[fld] = val
					}
				}
				filtered = append(filtered, pruned)
			} else {
				filtered = append(filtered, record)
			}
		}
	}

	// Output filtered records
	if extract {
		encoder := json.NewEncoder(os.Stdout)
		if pretty {
			encoder.SetIndent("", "  ")
		}
		return encoder.Encode(filtered)
	}

	if strings.ToLower(format) == "jsonl" {
		return parser.WriteJSONL(os.Stdout, filtered, pretty)
	}
	return parser.WriteJSON(os.Stdout, filtered, pretty)
}

func runFilter(cmd *cobra.Command, args []string) error {
	var filename string
	var field, operator, value string

	// Parse arguments
	if len(args) == 0 {
		// Reading from stdin, check for expression in flags
		filename = "-"
		if filterField == "" {
			return fmt.Errorf("when reading from stdin, provide filter expression or use --field, --op, --value flags")
		}
		field = filterField
		operator = filterOperator
		value = filterValue
	} else if len(args) == 1 {
		// One argument: could be filename or expression
		arg := args[0]

		// Check if it's an expression (contains operator)
		if expr := query.ParseFilterExpression(arg); expr != nil {
			// It's an expression, read from stdin
			filename = "-"
			field = expr.Field
			operator = expr.Operator
			value = expr.Value
		} else if filterField != "" {
			// It's a filename with flags
			filename = arg
			field = filterField
			operator = filterOperator
			value = filterValue
		} else {
			return fmt.Errorf("provide filter expression (e.g., age>28) or use --field, --op, --value flags")
		}
	} else if len(args) == 2 {
		// Two arguments: filename and expression
		filename = args[0]
		expr := query.ParseFilterExpression(args[1])
		if expr == nil {
			return fmt.Errorf("invalid filter expression: %s (use format: field>value)", args[1])
		}
		field = expr.Field
		operator = expr.Operator
		value = expr.Value
	} else {
		return fmt.Errorf("too many arguments")
	}

	return RunFilter(filename, field, operator, value, filterPretty, false, QuerySelect, filterFormat)
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
