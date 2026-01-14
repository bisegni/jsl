package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsl",
	Short: "JSON and JSONL query tool",
	Long: `jsl is a command-line tool for querying, filtering, and manipulating JSON and JSONL files.
It provides various commands to work with JSON data in bash environments.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(filterCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(validateCmd)
}
