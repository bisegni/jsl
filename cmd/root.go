package cmd

import (
	"os"

	"github.com/spf13/cobra"
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

		if len(args) == 0 {
			if hasStdin {
				// Data is being piped to stdin
				return RunQuery("-", QueryPath, QueryPretty)
			}
			return cmd.Help()
		}
		
		// One argument
		if len(args) == 1 {
			arg := args[0]
			// If we have stdin and arg looks like a path (starts with .), use it as path
			if hasStdin && len(arg) > 0 && arg[0] == '.' {
				return RunQuery("-", arg, QueryPretty)
			}
			// Otherwise it's a filename
			return RunQuery(arg, QueryPath, QueryPretty)
		}
		
		// Two arguments: filename and path
		filename := args[0]
		path := args[1]
		return RunQuery(filename, path, QueryPretty)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&QueryPath, "path", "p", ".", "Path to extract (e.g., .user.name)")
	rootCmd.PersistentFlags().BoolVar(&QueryPretty, "pretty", true, "Pretty print output")

	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(validateCmd)
}
