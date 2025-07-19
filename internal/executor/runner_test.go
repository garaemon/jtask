package executor

import (
	"os"
	"strings"
	"testing"

	"github.com/garaemon/jtask/internal/config"
)

func TestBuildCommand_Shell(t *testing.T) {
	task := &config.Task{
		Type:    "shell",
		Command: "echo hello",
		Args:    []string{"world"},
	}

	cmd := buildCommand(task)
	
	if cmd.Path != "/bin/sh" {
		t.Errorf("expected shell path to be /bin/sh, got %s", cmd.Path)
	}
	
	expectedArgs := []string{"/bin/sh", "-c", "echo hello world"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}
	
	for i, arg := range expectedArgs {
		if cmd.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, cmd.Args[i])
		}
	}
}

func TestBuildCommand_Process(t *testing.T) {
	task := &config.Task{
		Type:    "process",
		Command: "go",
		Args:    []string{"build", "-v"},
	}

	cmd := buildCommand(task)
	
	if !strings.HasSuffix(cmd.Path, "go") {
		t.Errorf("expected command path to end with go, got %s", cmd.Path)
	}
	
	expectedArgs := []string{"go", "build", "-v"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}
	
	for i, arg := range expectedArgs {
		if cmd.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, cmd.Args[i])
		}
	}
}

func TestBuildShellCommand_WithCustomShell(t *testing.T) {
	task := &config.Task{
		Type:    "shell",
		Command: "ls",
		Options: &config.TaskOptions{
			Shell: &config.ShellOptions{
				Executable: "/bin/bash",
				Args:       []string{"-c"},
			},
		},
	}

	cmd := buildShellCommand(task)
	
	if cmd.Path != "/bin/bash" {
		t.Errorf("expected shell path to be /bin/bash, got %s", cmd.Path)
	}
	
	expectedArgs := []string{"/bin/bash", "-c", "ls"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}
	
	for i, arg := range expectedArgs {
		if cmd.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, cmd.Args[i])
		}
	}
}

func TestBuildProcessCommand_NoArgs(t *testing.T) {
	task := &config.Task{
		Type:    "process",
		Command: "pwd",
	}

	cmd := buildProcessCommand(task)
	
	if !strings.HasSuffix(cmd.Path, "pwd") {
		t.Errorf("expected command path to end with pwd, got %s", cmd.Path)
	}
	
	expectedArgs := []string{"pwd"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}
	
	for i, arg := range expectedArgs {
		if cmd.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, cmd.Args[i])
		}
	}
}

func TestBuildEnvVars(t *testing.T) {
	envMap := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	envVars := buildEnvVars(envMap)
	
	if len(envVars) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(envVars))
	}
	
	found := make(map[string]bool)
	for _, env := range envVars {
		switch env {
		case "FOO=bar":
			found["FOO"] = true
		case "BAZ=qux":
			found["BAZ"] = true
		}
	}
	
	if !found["FOO"] || !found["BAZ"] {
		t.Errorf("expected env vars FOO=bar and BAZ=qux, got %v", envVars)
	}
}

func TestRunTask_UnsupportedType(t *testing.T) {
	task := &config.Task{
		Type:    "unsupported",
		Command: "echo hello",
	}

	err := RunTask(task, "/tmp", "")
	
	if err == nil {
		t.Error("expected error for unsupported task type")
	}
	
	if err.Error() != "unsupported task type: unsupported" {
		t.Errorf("expected error message about unsupported type, got %s", err.Error())
	}
}

func TestRunTask_EmptyCommand(t *testing.T) {
	task := &config.Task{
		Type:    "shell",
		Command: "",
	}

	err := RunTask(task, "/tmp", "")
	
	if err == nil {
		t.Error("expected error for empty command")
	}
	
	if err.Error() != "task command is empty" {
		t.Errorf("expected error message about empty command, got %s", err.Error())
	}
}

func TestSubstituteVariables(t *testing.T) {
	workspaceDir := "/home/user/project"
	
	task := &config.Task{
		Type:    "shell",
		Command: "ls ${workspaceFolder}",
		Args:    []string{"${workspaceFolder}/src", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}/build",
			Env: map[string]string{
				"PROJECT_ROOT": "${workspaceFolder}",
				"OTHER_VAR":    "value",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, "")
	
	if substituted.Command != "ls /home/user/project" {
		t.Errorf("expected command to be 'ls /home/user/project', got '%s'", substituted.Command)
	}
	
	expectedArgs := []string{"/home/user/project/src", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	if substituted.Options.Cwd != "/home/user/project/build" {
		t.Errorf("expected cwd to be '/home/user/project/build', got '%s'", substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["PROJECT_ROOT"] != "/home/user/project" {
		t.Errorf("expected PROJECT_ROOT to be '/home/user/project', got '%s'", substituted.Options.Env["PROJECT_ROOT"])
	}
	
	if substituted.Options.Env["OTHER_VAR"] != "value" {
		t.Errorf("expected OTHER_VAR to be 'value', got '%s'", substituted.Options.Env["OTHER_VAR"])
	}
}

func TestSubstituteVariables_WithFile(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/main.go"
	
	task := &config.Task{
		Type:    "shell",
		Command: "cat ${file}",
		Args:    []string{"${workspaceFolder}/${file}", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}",
			Env: map[string]string{
				"TARGET_FILE": "${file}",
				"PROJECT_ROOT": "${workspaceFolder}",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	if substituted.Command != "cat src/main.go" {
		t.Errorf("expected command to be 'cat src/main.go', got '%s'", substituted.Command)
	}
	
	expectedArgs := []string{"/home/user/project/src/main.go", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	if substituted.Options.Cwd != "/home/user/project" {
		t.Errorf("expected cwd to be '/home/user/project', got '%s'", substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["TARGET_FILE"] != "src/main.go" {
		t.Errorf("expected TARGET_FILE to be 'src/main.go', got '%s'", substituted.Options.Env["TARGET_FILE"])
	}
	
	if substituted.Options.Env["PROJECT_ROOT"] != "/home/user/project" {
		t.Errorf("expected PROJECT_ROOT to be '/home/user/project', got '%s'", substituted.Options.Env["PROJECT_ROOT"])
	}
}

func TestSubstituteVariables_WithCwd(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/main.go"
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${cwd}",
		Args:    []string{"${cwd}/build", "test"},
		Options: &config.TaskOptions{
			Cwd: "${cwd}/output",
			Env: map[string]string{
				"CURRENT_DIR": "${cwd}",
				"BUILD_DIR":   "${cwd}/build",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	// Get expected cwd value (should be the current working directory)
	expectedCwd, err := os.Getwd()
	if err != nil {
		expectedCwd = ""
	}
	
	expectedCommand := "echo " + expectedCwd
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{expectedCwd + "/build", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := expectedCwd + "/output"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["CURRENT_DIR"] != expectedCwd {
		t.Errorf("expected CURRENT_DIR to be '%s', got '%s'", expectedCwd, substituted.Options.Env["CURRENT_DIR"])
	}
	
	expectedBuildDir := expectedCwd + "/build"
	if substituted.Options.Env["BUILD_DIR"] != expectedBuildDir {
		t.Errorf("expected BUILD_DIR to be '%s', got '%s'", expectedBuildDir, substituted.Options.Env["BUILD_DIR"])
	}
}