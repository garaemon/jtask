package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/garaemon/jtask/internal/config"
)

func executeTask(task *config.Task) error {
	if task.Type != "shell" && task.Type != "process" {
		return fmt.Errorf("unsupported task type: %s", task.Type)
	}

	if task.Command == "" {
		return fmt.Errorf("task command is empty")
	}

	cmd := buildCommand(task)
	
	if task.Options != nil && task.Options.Cwd != "" {
		cmd.Dir = task.Options.Cwd
	}

	if task.Options != nil && task.Options.Env != nil {
		cmd.Env = append(os.Environ(), buildEnvVars(task.Options.Env)...)
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

func RunTask(task *config.Task) error {
	return executeTask(task)
}