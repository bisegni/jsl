package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/engine"
	"github.com/bisegni/jsl/pkg/query"
	"github.com/spf13/cobra"
)

var (
	QueryPath       string
	QueryPretty     bool
	QueryExtract    bool
	QuerySelect     []string
	InteractiveMode bool
)

var rootCmd = &cobra.Command{
	Use:   "jsl [file|JSON] [path]",
	Short: "JSON and JSONL query tool",
	Long: `jsl is a command-line tool for querying, filtering, and manipulating JSON and JSONL files.
If no command is provided, it defaults to querying the specified file.

Supports:
  - File paths: jsl data.json .user.name
  - Stdin: cat data.json | jsl .user.name  (or use "-" as filename)
  - Inline JSON: jsl '{"name":"Alice"}' .name

Examples:
  jsl data.json .user.name
  cat data.json | jsl .user.name
  echo '{"name":"Alice"}' | jsl .name
  jsl '{"name":"Alice","age":30}' .name
  jsl stats data.jsonl`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		hasStdin := (stat.Mode() & os.ModeCharDevice) == 0

		if InteractiveMode {
			var filename string
			if len(args) > 0 {
				filename = args[0]
			} else if hasStdin {
				filename = "-"
			} else {
				return fmt.Errorf("interactive mode requires a file or stdin input")
			}
			return RunInteractive(filename)
		}

		var filename, expression string

		if len(args) == 0 {
			if hasStdin {
				filename = "-"
				expression = QueryPath
			} else {
				return cmd.Help()
			}
		} else if len(args) == 1 {
			arg := args[0]
			if hasStdin {
				filename = "-"
				expression = arg
			} else {
				// If not stdin, it could be a filename (default query) or
				// if we have flags, maybe an expression?
				// Usually with 1 arg and no stdin, it's a filename.
				filename = arg
				expression = QueryPath
			}
		} else {
			// Two arguments: filename and (path or expression)
			filename = args[0]
			expression = args[1]
		}

		// Intelligent routing
		// Check if it's a SQL-like query
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(expression)), "SELECT") {
			q, err := engine.ParseQuery(expression)
			if err != nil {
				return fmt.Errorf("failed to parse query: %w", err)
			}

			// Create Input Table
			inputTable := database.NewJSONTable(filename)

			// Execute
			// Execute
			executor := engine.NewExecutor()
			executor.Pretty = QueryPretty
			return executor.Execute(q, inputTable, os.Stdout)
		}

		if query.IsFilterExpression(expression) {
			expr := query.ParseFilterExpression(expression)
			if expr != nil {
				return RunFilter(filename, expr.Field, expr.Operator, expr.Value, QueryPretty, QueryExtract, QuerySelect, "json")
			}
		}

		return RunQuery(filename, expression, QueryPretty, QueryExtract, QuerySelect)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&QueryPath, "path", "p", ".", "Path to extract (e.g., .user.name)")
	rootCmd.PersistentFlags().BoolVar(&QueryPretty, "pretty", false, "Pretty print output")
	rootCmd.PersistentFlags().BoolVarP(&QueryExtract, "extract", "e", false, "Extract mode (flattened line-by-line output)")
	rootCmd.PersistentFlags().StringSliceVarP(&QuerySelect, "select", "s", []string{}, "Select specific fields to include in output (e.g., value,metadata)")
	rootCmd.PersistentFlags().BoolVarP(&InteractiveMode, "interactive", "i", false, "Interactive REPL mode")

	// Subcommands that still make sense as separate actions
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(validateCmd)
}
