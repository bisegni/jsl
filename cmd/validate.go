package cmd

import (
	"fmt"

	"github.com/bisegni/jsl/pkg/parser"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [file|-]",
	Short: "Validate JSON/JSONL file syntax",
	Long: `Validate that a JSON or JSONL file has correct syntax.
	
Supports:
  - File paths: jsl validate data.json
  - Stdin: cat data.json | jsl validate

Examples:
  jsl validate data.json
  jsl validate data.jsonl
  cat data.json | jsl validate`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
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
		fmt.Printf("❌ Validation failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Valid %s file with %d record(s)\n", getFormat(p.IsJSONL()), len(records))
	return nil
}
