package cmd

import (
	"os"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	formatPretty bool
	formatOutput string
)

var formatCmd = &cobra.Command{
	Use:   "format [file|-]",
	Short: "Format and pretty-print JSON/JSONL file",
	Long: `Format and pretty-print a JSON or JSONL file.
	
Supports:
  - File paths: jsl format data.json
  - Stdin: cat data.json | jsl format (or use "-")

Examples:
  jsl format data.json
  jsl format data.jsonl --output jsonl
  cat data.json | jsl format
  echo '{"name":"Alice"}' | jsl format`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFormat,
}

func init() {
	formatCmd.Flags().BoolVar(&formatPretty, "pretty", true, "Pretty print output")
	formatCmd.Flags().StringVarP(&formatOutput, "output", "o", "", "Output format (json or jsonl, auto-detect if not specified)")
}

func runFormat(cmd *cobra.Command, args []string) error {
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

	// Determine output format
	outputFormat := formatOutput
	if outputFormat == "" {
		// Auto-detect from input
		if p.IsJSONL() {
			outputFormat = "jsonl"
		} else {
			outputFormat = "json"
		}
	}

	// Output formatted records
	if outputFormat == "jsonl" {
		return parser.WriteJSONL(os.Stdout, records, formatPretty)
	}
	return parser.WriteJSON(os.Stdout, records, formatPretty)
}
