package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsl [file] [path]",
	Short: "JSON and JSONL query tool",
	Long: `jsl is a command-line tool for querying, filtering, and manipulating JSON and JSONL files.
If no command is provided, it defaults to querying the specified file.

Examples:
  jsl data.json .user.name
  jsl stats data.jsonl`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		filename := args[0]
		path := QueryPath
		if len(args) > 1 {
			path = args[1]
		}
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
