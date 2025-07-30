package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/garaemon/tasks-json-cli/internal/config"
	"github.com/garaemon/tasks-json-cli/internal/discovery"
	"github.com/spf13/cobra"
)

var (
	groupFilter string
	typeFilter  string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tasks",
	Long:  `List all available tasks from the tasks.json file.`,
	RunE:  runListCommand,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&groupFilter, "group", "", "filter by group (build, test)")
	listCmd.Flags().StringVar(&typeFilter, "type", "", "filter by type (shell, process)")
}

func runListCommand(cmd *cobra.Command, args []string) error {
	tasksPath, err := discovery.FindTasksFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to find tasks file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Loading tasks from: %s\n", tasksPath)
	}

	tasks, err := config.LoadTasks(tasksPath)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	filteredTasks := filterTasks(tasks, groupFilter, typeFilter)

	if len(filteredTasks) == 0 {
		if !quiet {
			fmt.Println("No tasks found")
		}
		return nil
	}

	printTasks(filteredTasks)
	return nil
}

func filterTasks(tasks []config.Task, groupFilter, typeFilter string) []config.Task {
	var filtered []config.Task

	for _, task := range tasks {
		if groupFilter != "" && task.GetGroupKind() != groupFilter {
			continue
		}
		if typeFilter != "" && task.Type != typeFilter {
			continue
		}
		filtered = append(filtered, task)
	}

	return filtered
}

func printTasks(tasks []config.Task) {
	if quiet {
		for _, task := range tasks {
			fmt.Println(task.Label)
		}
		return
	}

	fmt.Printf("%-20s %-8s %-8s %s\n", "LABEL", "TYPE", "GROUP", "COMMAND")
	fmt.Println(strings.Repeat("-", 60))

	for _, task := range tasks {
		group := task.GetGroupKind()
		if group == "" {
			group = "-"
		}

		command := task.Command
		if len(task.Args) > 0 {
			command += " " + strings.Join(task.Args, " ")
		}

		if len(command) > 30 {
			command = command[:27] + "..."
		}

		fmt.Printf("%-20s %-8s %-8s %s\n", task.Label, task.Type, group, command)
	}
}