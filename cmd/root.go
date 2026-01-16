package cmd

import (
	"os"

	"github.com/bisegni/jsl/pkg/query"
	"github.com/spf13/cobra"
)

var (
	QueryPath    string
	QueryPretty  bool
	QueryExtract bool
	QuerySelect  []string
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
	rootCmd.PersistentFlags().BoolVar(&QueryPretty, "pretty", true, "Pretty print output")
	rootCmd.PersistentFlags().BoolVarP(&QueryExtract, "extract", "e", false, "Extract mode (flattened line-by-line output)")
	rootCmd.PersistentFlags().StringSliceVarP(&QuerySelect, "select", "s", []string{}, "Select specific fields to include in output (e.g., value,metadata)")

	// Subcommands that still make sense as separate actions
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(validateCmd)
}
