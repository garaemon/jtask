package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/garaemon/jtask/internal/config"
	"github.com/garaemon/jtask/internal/discovery"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <task-name>",
	Short: "Show task details",
	Long:  `Show detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInfoCommand,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfoCommand(cmd *cobra.Command, args []string) error {
	taskName := args[0]
	
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

	task := findTaskByName(tasks, taskName)
	if task == nil {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	printTaskInfo(task, tasksPath)
	return nil
}

func findTaskByName(tasks []config.Task, name string) *config.Task {
	for _, task := range tasks {
		if task.Label == name {
			return &task
		}
	}
	return nil
}

func printTaskInfo(task *config.Task, tasksPath string) {
	if quiet {
		printTaskInfoQuiet(task)
		return
	}

	fmt.Printf("Task: %s\n", task.Label)
	fmt.Println(strings.Repeat("=", len(task.Label)+6))
	fmt.Println()

	// Basic information
	fmt.Printf("Type:     %s\n", task.Type)
	fmt.Printf("Command:  %s\n", task.Command)
	
	if len(task.Args) > 0 {
		fmt.Printf("Args:     %s\n", strings.Join(task.Args, " "))
	}

	group := task.GetGroupKind()
	if group != "" {
		fmt.Printf("Group:    %s\n", group)
	}


	// Options
	if task.Options != nil {
		fmt.Println()
		fmt.Println("Options:")
		
		if task.Options.Cwd != "" {
			fmt.Printf("  Working Directory: %s\n", task.Options.Cwd)
		}
		
		if len(task.Options.Env) > 0 {
			fmt.Println("  Environment Variables:")
			for key, value := range task.Options.Env {
				fmt.Printf("    %s=%s\n", key, value)
			}
		}
		
		if task.Options.Shell != nil {
			fmt.Println("  Shell Options:")
			if task.Options.Shell.Executable != "" {
				fmt.Printf("    Executable: %s\n", task.Options.Shell.Executable)
			}
			if len(task.Options.Shell.Args) > 0 {
				fmt.Printf("    Args: %s\n", strings.Join(task.Options.Shell.Args, " "))
			}
		}
	}

	// Dependencies
	dependsOn := getDependsOnAsStringSlice(task.DependsOn)
	if len(dependsOn) > 0 {
		fmt.Println()
		fmt.Printf("Depends On: %s\n", strings.Join(dependsOn, ", "))
	}

	// Problem matcher
	if task.ProblemMatcher != nil {
		fmt.Println()
		fmt.Printf("Problem Matcher: %v\n", task.ProblemMatcher)
	}

	// Verbose information
	if verbose {
		fmt.Println()
		fmt.Println("Additional Information:")
		fmt.Printf("  Source File: %s\n", tasksPath)
		fmt.Println("  Note: Use 'jtask run' with --file flag to see variable substitution")
	}
}

func printTaskInfoQuiet(task *config.Task) {
	fmt.Printf("%s\t%s\t%s", task.Label, task.Type, task.Command)
	if len(task.Args) > 0 {
		fmt.Printf(" %s", strings.Join(task.Args, " "))
	}
	fmt.Println()
}

func getDependsOnAsStringSlice(dependsOn interface{}) []string {
	if dependsOn == nil {
		return nil
	}
	
	switch deps := dependsOn.(type) {
	case string:
		return []string{deps}
	case []string:
		return deps
	case []interface{}:
		var result []string
		for _, dep := range deps {
			if s, ok := dep.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	
	return nil
}