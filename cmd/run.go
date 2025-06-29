package cmd

import (
	"fmt"

	"github.com/garaemon/jtask/internal/config"
	"github.com/garaemon/jtask/internal/discovery"
	"github.com/garaemon/jtask/internal/executor"
	"github.com/spf13/cobra"
)

var dryRun bool

var runCommand = &cobra.Command{
	Use:   "run <task-name>",
	Short: "Execute specified task",
	Long:  `Execute a task defined in the tasks.json file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  executeRunCommand,
}

func executeRunCommand(cmd *cobra.Command, args []string) error {
	taskName := args[0]

	tasksFilePath, err := discovery.FindTasksFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to find tasks file: %w", err)
	}

	if verbose {
		fmt.Printf("Using tasks file: %s\n", tasksFilePath)
	}

	tasks, err := config.LoadTasks(tasksFilePath)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	var targetTask *config.Task
	for i := range tasks {
		if tasks[i].Label == taskName {
			targetTask = &tasks[i]
			break
		}
	}

	if targetTask == nil {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	if dryRun {
		fmt.Printf("Would execute task: %s\n", targetTask.Label)
		fmt.Printf("  Type: %s\n", targetTask.Type)
		fmt.Printf("  Command: %s\n", targetTask.Command)
		if len(targetTask.Args) > 0 {
			fmt.Printf("  Args: %v\n", targetTask.Args)
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Executing task: %s\n", targetTask.Label)
	}

	return executor.RunTask(targetTask)
}

func init() {
	runCommand.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be executed without running")
	rootCmd.AddCommand(runCommand)
}