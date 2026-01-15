package cmd

import (
	"os"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	convertOutput string
	convertPretty bool
)

var convertCmd = &cobra.Command{
	Use:   "convert [file|-]",
	Short: "Convert between JSON and JSONL formats",
	Long: `Convert a file between JSON and JSONL formats.
	
Supports:
  - File paths: jsl convert data.json --to jsonl
  - Stdin: cat data.json | jsl convert --to jsonl

Examples:
  jsl convert data.json --to jsonl
  jsl convert data.jsonl --to json
  cat data.json | jsl convert --to jsonl
  echo '{"name":"Alice"}' | jsl convert --to jsonl`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().StringVarP(&convertOutput, "to", "t", "", "Target format (json or jsonl)")
	convertCmd.Flags().BoolVar(&convertPretty, "pretty", true, "Pretty print output")
	convertCmd.MarkFlagRequired("to")
}

func runConvert(cmd *cobra.Command, args []string) error {
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

	// Output in target format
	if convertOutput == "jsonl" {
		return parser.WriteJSONL(os.Stdout, records, convertPretty)
	}
	return parser.WriteJSON(os.Stdout, records, convertPretty)
}
