package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/garaemon/tasks-json-cli/internal/config"
)

func executeTask(task *config.Task, workspaceDir string, file string) error {
	supportedTypes := []string{"shell", "process", "npm", "typescript"}
	isSupported := false
	for _, t := range supportedTypes {
		if task.Type == t {
			isSupported = true
			break
		}
	}
	
	if !isSupported {
		return fmt.Errorf("unsupported task type: %s", task.Type)
	}

	// Check command requirements for specific task types
	if (task.Type == "shell" || task.Type == "process") && task.Command == "" {
		return fmt.Errorf("task command is empty")
	}

	// Apply variable substitution
	substitutedTask := substituteVariables(task, workspaceDir, file)
	
	// Build command based on task type
	cmd, err := buildCommandForTaskType(substitutedTask, workspaceDir)
	if err != nil {
		return err
	}
	
	if substitutedTask.Options != nil && substitutedTask.Options.Cwd != "" {
		cmd.Dir = substitutedTask.Options.Cwd
	}

	if substitutedTask.Options != nil && substitutedTask.Options.Env != nil {
		cmd.Env = append(os.Environ(), buildEnvVars(substitutedTask.Options.Env)...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func buildCommandForTaskType(task *config.Task, workspaceDir string) (*exec.Cmd, error) {
	switch task.Type {
	case "shell":
		return buildShellCommand(task), nil
	case "process":
		return buildProcessCommand(task), nil
	case "npm":
		return buildNpmCommand(task, workspaceDir)
	case "typescript":
		return buildTypescriptCommand(task, workspaceDir)
	default:
		return nil, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}

func buildCommand(task *config.Task) *exec.Cmd {
	if task.Type == "shell" {
		return buildShellCommand(task)
	}
	return buildProcessCommand(task)
}

func buildShellCommand(task *config.Task) *exec.Cmd {
	shell := "/bin/sh"
	shellArgs := []string{"-c"}

	if task.Options != nil && task.Options.Shell != nil {
		if task.Options.Shell.Executable != "" {
			shell = task.Options.Shell.Executable
		}
		if len(task.Options.Shell.Args) > 0 {
			shellArgs = task.Options.Shell.Args
		}
	}

	commandLine := task.Command
	if len(task.Args) > 0 {
		commandLine += " " + strings.Join(task.Args, " ")
	}

	args := append(shellArgs, commandLine)
	return exec.Command(shell, args...)
}

func buildProcessCommand(task *config.Task) *exec.Cmd {
	if len(task.Args) == 0 {
		return exec.Command(task.Command)
	}
	return exec.Command(task.Command, task.Args...)
}

func buildNpmCommand(task *config.Task, workspaceDir string) (*exec.Cmd, error) {
	if task.Script == "" {
		return nil, fmt.Errorf("npm task requires 'script' field")
	}
	
	args := []string{"run", task.Script}
	cmd := exec.Command("npm", args...)
	
	// Set working directory to task path or workspace directory
	if task.Path != "" {
		if filepath.IsAbs(task.Path) {
			cmd.Dir = task.Path
		} else {
			cmd.Dir = filepath.Join(workspaceDir, task.Path)
		}
	} else {
		cmd.Dir = workspaceDir
	}
	
	return cmd, nil
}

func buildTypescriptCommand(task *config.Task, workspaceDir string) (*exec.Cmd, error) {
	var args []string
	
	// Add tsconfig if specified
	if task.TSConfig != "" {
		args = append(args, "-p", task.TSConfig)
	}
	
	// Add option if specified (e.g., "watch")
	if task.Option != "" {
		if task.Option == "watch" {
			args = append(args, "--watch")
		} else {
			args = append(args, task.Option)
		}
	}
	
	cmd := exec.Command("tsc", args...)
	cmd.Dir = workspaceDir
	
	return cmd, nil
}

func buildEnvVars(envMap map[string]string) []string {
	var envVars []string
	for key, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}
	return envVars
}

// substituteEnvVariables replaces ${env:VARNAME} patterns with environment variable values
func substituteEnvVariables(text string) string {
	envVarPattern := regexp.MustCompile(`\$\{env:([^}]+)\}`)
	return envVarPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name from ${env:VARNAME}
		varName := envVarPattern.FindStringSubmatch(match)[1]
		// Get environment variable value, return empty string if not found
		return os.Getenv(varName)
	})
}

func substituteVariables(task *config.Task, workspaceDir string, file string) *config.Task {
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
	
	// Get file directory for ${fileDirname} variable
	fileDirname := ""
	if file != "" {
		fileDirname = filepath.Dir(file)
	}
	
	// Get file extension for ${fileExtname} variable
	fileExtname := ""
	if file != "" {
		fileExtname = filepath.Ext(file)
	}
	
	// Get workspace folder of the current file for ${fileWorkspaceFolder} variable
	fileWorkspaceFolder := ""
	if file != "" {
		// If file is absolute path, check if it's within workspace
		if filepath.IsAbs(file) {
			// Check if file is within the workspace directory
			rel, err := filepath.Rel(workspaceDir, file)
			if err == nil && !strings.HasPrefix(rel, "..") {
				fileWorkspaceFolder = workspaceDir
			}
		} else {
			// If file is relative, assume it's relative to workspace
			fileWorkspaceFolder = workspaceDir
		}
	}
	
	// Get relative file path for ${relativeFile} variable
	relativeFile := ""
	if file != "" {
		if filepath.IsAbs(file) {
			// If file is absolute, get relative path from workspace
			rel, err := filepath.Rel(workspaceDir, file)
			if err == nil && !strings.HasPrefix(rel, "..") {
				relativeFile = rel
			}
		} else {
			// If file is already relative, use as is
			relativeFile = file
		}
	}
	
	// Get relative file directory path for ${relativeFileDirname} variable
	relativeFileDirname := ""
	if relativeFile != "" {
		relativeFileDirname = filepath.Dir(relativeFile)
	}
	
	// Create a copy of the task to avoid modifying the original
	substituted := *task
	
	// Replace variables in command
	substituted.Command = strings.ReplaceAll(task.Command, "${workspaceFolder}", workspaceDir)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${workspaceFolderBasename}", workspaceFolderBasename)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${file}", file)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileBasename}", fileBasename)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileBasenameNoExtension}", fileBasenameNoExtension)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileDirname}", fileDirname)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileExtname}", fileExtname)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${fileWorkspaceFolder}", fileWorkspaceFolder)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${relativeFile}", relativeFile)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${relativeFileDirname}", relativeFileDirname)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${cwd}", cwd)
	substituted.Command = strings.ReplaceAll(substituted.Command, "${pathSeparator}", pathSeparator)
	substituted.Command = substituteEnvVariables(substituted.Command)
	
	// Replace variables in args
	if len(task.Args) > 0 {
		substituted.Args = make([]string, len(task.Args))
		for i, arg := range task.Args {
			substituted.Args[i] = strings.ReplaceAll(arg, "${workspaceFolder}", workspaceDir)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${workspaceFolderBasename}", workspaceFolderBasename)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${file}", file)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileBasename}", fileBasename)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileBasenameNoExtension}", fileBasenameNoExtension)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileDirname}", fileDirname)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileExtname}", fileExtname)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${fileWorkspaceFolder}", fileWorkspaceFolder)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${relativeFile}", relativeFile)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${relativeFileDirname}", relativeFileDirname)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${cwd}", cwd)
			substituted.Args[i] = strings.ReplaceAll(substituted.Args[i], "${pathSeparator}", pathSeparator)
			substituted.Args[i] = substituteEnvVariables(substituted.Args[i])
		}
	}
	
	// Replace variables in options if present
	if task.Options != nil {
		substituted.Options = &config.TaskOptions{}
		*substituted.Options = *task.Options
		
		if task.Options.Cwd != "" {
			substituted.Options.Cwd = strings.ReplaceAll(task.Options.Cwd, "${workspaceFolder}", workspaceDir)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${workspaceFolderBasename}", workspaceFolderBasename)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${file}", file)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${fileBasename}", fileBasename)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${fileBasenameNoExtension}", fileBasenameNoExtension)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${fileDirname}", fileDirname)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${fileExtname}", fileExtname)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${fileWorkspaceFolder}", fileWorkspaceFolder)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${relativeFile}", relativeFile)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${relativeFileDirname}", relativeFileDirname)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${cwd}", cwd)
			substituted.Options.Cwd = strings.ReplaceAll(substituted.Options.Cwd, "${pathSeparator}", pathSeparator)
			substituted.Options.Cwd = substituteEnvVariables(substituted.Options.Cwd)
		}
		
		if task.Options.Env != nil {
			substituted.Options.Env = make(map[string]string)
			for key, value := range task.Options.Env {
				substituted.Options.Env[key] = strings.ReplaceAll(value, "${workspaceFolder}", workspaceDir)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${workspaceFolderBasename}", workspaceFolderBasename)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${file}", file)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${fileBasename}", fileBasename)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${fileBasenameNoExtension}", fileBasenameNoExtension)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${fileDirname}", fileDirname)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${fileExtname}", fileExtname)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${fileWorkspaceFolder}", fileWorkspaceFolder)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${relativeFile}", relativeFile)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${relativeFileDirname}", relativeFileDirname)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${cwd}", cwd)
				substituted.Options.Env[key] = strings.ReplaceAll(substituted.Options.Env[key], "${pathSeparator}", pathSeparator)
				substituted.Options.Env[key] = substituteEnvVariables(substituted.Options.Env[key])
			}
		}
	}
	
	// Replace variables in TypeScript-specific fields
	if task.TSConfig != "" {
		substituted.TSConfig = strings.ReplaceAll(task.TSConfig, "${workspaceFolder}", workspaceDir)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${workspaceFolderBasename}", workspaceFolderBasename)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${file}", file)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${fileBasename}", fileBasename)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${fileBasenameNoExtension}", fileBasenameNoExtension)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${fileDirname}", fileDirname)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${fileExtname}", fileExtname)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${fileWorkspaceFolder}", fileWorkspaceFolder)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${relativeFile}", relativeFile)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${relativeFileDirname}", relativeFileDirname)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${cwd}", cwd)
		substituted.TSConfig = strings.ReplaceAll(substituted.TSConfig, "${pathSeparator}", pathSeparator)
		substituted.TSConfig = substituteEnvVariables(substituted.TSConfig)
	}
	
	// Replace variables in NPM-specific fields
	if task.Path != "" {
		substituted.Path = strings.ReplaceAll(task.Path, "${workspaceFolder}", workspaceDir)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${workspaceFolderBasename}", workspaceFolderBasename)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${file}", file)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${fileBasename}", fileBasename)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${fileBasenameNoExtension}", fileBasenameNoExtension)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${fileDirname}", fileDirname)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${fileExtname}", fileExtname)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${fileWorkspaceFolder}", fileWorkspaceFolder)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${relativeFile}", relativeFile)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${relativeFileDirname}", relativeFileDirname)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${cwd}", cwd)
		substituted.Path = strings.ReplaceAll(substituted.Path, "${pathSeparator}", pathSeparator)
		substituted.Path = substituteEnvVariables(substituted.Path)
	}
	
	return &substituted
}

func RunTask(task *config.Task, workspaceDir string, file string) error {
	return executeTask(task, workspaceDir, file)
}

func RunTaskWithDependencies(task *config.Task, allTasks []config.Task, workspaceDir string, file string) error {
	resolver := NewDependencyResolver(allTasks)
	
	executionOrder, err := resolver.ResolveExecutionOrder(task.Label)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}
	
	for _, t := range executionOrder {
		err := executeTask(t, workspaceDir, file)
		if err != nil {
			return fmt.Errorf("failed to execute task '%s': %w", t.Label, err)
		}
	}
	
	return nil
}