package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/garaemon/jtask/internal/config"
	"github.com/garaemon/jtask/internal/discovery"
	"github.com/garaemon/jtask/internal/executor"
	"github.com/spf13/cobra"
)

var dryRun bool
var workspaceFolder string
var file string

var runCommand = &cobra.Command{
	Use:   "run <task-name>",
	Short: "Execute specified task",
	Long:  `Execute a task defined in the tasks.json file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  executeRunCommand,
}

func executeRunCommand(cmd *cobra.Command, args []string) error {
	taskName := args[0]

	// Determine workspace folder
	var workspaceDir string
	if workspaceFolder != "" {
		workspaceDir = workspaceFolder
	} else {
		// Default to git root
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		gitRoot, err := discovery.FindGitRoot(currentDir)
		if err != nil {
			// Fall back to current directory if git root not found
			workspaceDir = currentDir
		} else {
			workspaceDir = gitRoot
		}
	}

	if verbose {
		fmt.Printf("Workspace folder: %s\n", workspaceDir)
	}

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
		// Apply variable substitution for dry-run display
		substitutedTask := substituteVariablesForDryRun(targetTask, workspaceDir, file)
		
		fmt.Printf("Would execute task: %s\n", substitutedTask.Label)
		fmt.Printf("  Type: %s\n", substitutedTask.Type)
		fmt.Printf("  Command: %s\n", substitutedTask.Command)
		if len(substitutedTask.Args) > 0 {
			fmt.Printf("  Args: %v\n", substitutedTask.Args)
		}
		return nil
	}

	if !quiet {
		fmt.Printf("Executing task: %s\n", targetTask.Label)
	}

	return executor.RunTask(targetTask, workspaceDir, file)
}

func substituteVariablesForDryRun(task *config.Task, workspaceDir string, file string) *config.Task {
	// Create a copy of the task to avoid modifying the original
	substituted := *task
	
	// Replace ${workspaceFolder} in command
	substituted.Command = strings.ReplaceAll(task.Command, "${workspaceFolder}", workspaceDir)
	// Replace ${file} in command
	substituted.Command = strings.ReplaceAll(substituted.Command, "${file}", file)
	
	// Replace ${workspaceFolder} and ${file} in args
	if len(task.Args) > 0 {
		substituted.Args = make([]string, len(task.Args))
		for i, arg := range task.Args {
			substituted.Args[i] = strings.ReplaceAll(arg, "${workspaceFolder}", workspaceDir)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${file}", file)
		}
	}
	
	return &substituted
}

func init() {
	runCommand.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be executed without running")
	runCommand.Flags().StringVar(&workspaceFolder, "workspace-folder", "", "workspace folder path (defaults to git root)")
	runCommand.Flags().StringVar(&file, "file", "", "file path to replace ${file} variable")
	rootCmd.AddCommand(runCommand)
}