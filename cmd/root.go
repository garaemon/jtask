package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
	quiet      bool
)

var rootCmd = &cobra.Command{
	Use:   "tasks-json-cli",
	Short: "Execute VS Code tasks from command line",
	Long: `tasks-json-cli is a CLI tool that executes tasks defined in VS Code's tasks.json configuration files.
It allows you to run VS Code tasks directly from the command line without needing the editor.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "specify tasks.json file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "minimal output")
}