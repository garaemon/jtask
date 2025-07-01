package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/garaemon/jtask/internal/config"
)

func executeTask(task *config.Task, workspaceDir string) error {
	if task.Type != "shell" && task.Type != "process" {
		return fmt.Errorf("unsupported task type: %s", task.Type)
	}

	if task.Command == "" {
		return fmt.Errorf("task command is empty")
	}

	// Apply variable substitution
	substitutedTask := substituteVariables(task, workspaceDir)
	
	cmd := buildCommand(substitutedTask)
	
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

func buildEnvVars(envMap map[string]string) []string {
	var envVars []string
	for key, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}
	return envVars
}

func substituteVariables(task *config.Task, workspaceDir string) *config.Task {
	// Create a copy of the task to avoid modifying the original
	substituted := *task
	
	// Replace ${workspaceFolder} in command
	substituted.Command = strings.ReplaceAll(task.Command, "${workspaceFolder}", workspaceDir)
	
	// Replace ${workspaceFolder} in args
	if len(task.Args) > 0 {
		substituted.Args = make([]string, len(task.Args))
		for i, arg := range task.Args {
			substituted.Args[i] = strings.ReplaceAll(arg, "${workspaceFolder}", workspaceDir)
		}
	}
	
	// Replace ${workspaceFolder} in options if present
	if task.Options != nil {
		substituted.Options = &config.TaskOptions{}
		*substituted.Options = *task.Options
		
		if task.Options.Cwd != "" {
			substituted.Options.Cwd = strings.ReplaceAll(task.Options.Cwd, "${workspaceFolder}", workspaceDir)
		}
		
		if task.Options.Env != nil {
			substituted.Options.Env = make(map[string]string)
			for key, value := range task.Options.Env {
				substituted.Options.Env[key] = strings.ReplaceAll(value, "${workspaceFolder}", workspaceDir)
			}
		}
	}
	
	return &substituted
}

func RunTask(task *config.Task, workspaceDir string) error {
	return executeTask(task, workspaceDir)
}