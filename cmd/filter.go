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
		if expr := parseFilterExpression(arg); expr != nil {
			// It's an expression, read from stdin
			filename = "-"
			field = expr.field
			operator = expr.operator
			value = expr.value
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
		expr := parseFilterExpression(args[1])
		if expr == nil {
			return fmt.Errorf("invalid filter expression: %s (use format: field>value)", args[1])
		}
		field = expr.field
		operator = expr.operator
		value = expr.value
	} else {
		return fmt.Errorf("too many arguments")
	}

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
			filtered = append(filtered, record)
		}
	}

	// Output filtered records
	if strings.ToLower(filterFormat) == "jsonl" {
		return parser.WriteJSONL(os.Stdout, filtered, filterPretty)
	}
	return parser.WriteJSON(os.Stdout, filtered, filterPretty)
}

type filterExpr struct {
	field    string
	operator string
	value    string
}

// parseFilterExpression parses expressions like "age>28", "name=john", "status!=active"
func parseFilterExpression(expr string) *filterExpr {
	// Try to find operator in the expression
	operators := []string{">=", "<=", "!=", "~=", ">", "<", "="}
	
	for _, op := range operators {
		if idx := strings.Index(expr, op); idx > 0 {
			field := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+len(op):])
			
			if field != "" && value != "" {
				// Convert ~= to contains for internal representation
				internalOp := op
				if op == "~=" {
					internalOp = "contains"
				}
				return &filterExpr{
					field:    field,
					operator: internalOp,
					value:    value,
				}
			}
		}
	}
	
	return nil
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
