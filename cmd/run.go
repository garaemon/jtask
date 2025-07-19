package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// substituteEnvVariablesForDryRun replaces ${env:VARNAME} patterns with environment variable values
func substituteEnvVariablesForDryRun(text string) string {
	envVarPattern := regexp.MustCompile(`\$\{env:([^}]+)\}`)
	return envVarPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name from ${env:VARNAME}
		varName := envVarPattern.FindStringSubmatch(match)[1]
		// Get environment variable value, return empty string if not found
		return os.Getenv(varName)
	})
}

func substituteVariablesForDryRun(task *config.Task, workspaceDir string, file string) *config.Task {
	// Get current working directory for ${cwd} variable
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to empty string if we can't get the current directory
		cwd = ""
	}
	
	// Get OS-specific path separator for ${pathSeparator} variable
	pathSeparator := string(filepath.Separator)
	
	// Get workspace folder basename for ${workspaceFolderBasename} variable
	workspaceFolderBasename := filepath.Base(workspaceDir)
	
	// Get file basename for ${fileBasename} variable
	fileBasename := ""
	if file != "" {
		fileBasename = filepath.Base(file)
	}
	
	// Get file basename without extension for ${fileBasenameNoExtension} variable
	fileBasenameNoExtension := ""
	if file != "" {
		basename := filepath.Base(file)
		ext := filepath.Ext(basename)
		fileBasenameNoExtension = strings.TrimSuffix(basename, ext)
	}
	
	// Create a copy of the task to avoid modifying the original
	substituted := *task
	
	// Replace variables in command
	substituted.Command = strings.ReplaceAll(task.Command, "${workspaceFolder}", workspaceDir)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${workspaceFolderBasename}", workspaceFolderBasename)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${file}", file)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileBasename}", fileBasename)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileBasenameNoExtension}", fileBasenameNoExtension)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${cwd}", cwd)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${pathSeparator}", pathSeparator)
	substituted.Command = substituteEnvVariablesForDryRun(substituted.Command)
	
	// Replace variables in args
	if len(task.Args) > 0 {
		substituted.Args = make([]string, len(task.Args))
		for i, arg := range task.Args {
			substituted.Args[i] = strings.ReplaceAll(arg, "${workspaceFolder}", workspaceDir)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${workspaceFolderBasename}", workspaceFolderBasename)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${file}", file)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileBasename}", fileBasename)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileBasenameNoExtension}", fileBasenameNoExtension)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${cwd}", cwd)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${pathSeparator}", pathSeparator)
			substituted.Args[i] = substituteEnvVariablesForDryRun(substituted.Args[i])
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