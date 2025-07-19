package executor

import (
	"os"
	"path/filepath"
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

func TestSubstituteVariables_WithPathSeparator(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/main.go"
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${pathSeparator}",
		Args:    []string{"path1${pathSeparator}path2", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}${pathSeparator}build",
			Env: map[string]string{
				"PATH_SEP":   "${pathSeparator}",
				"BUILD_PATH": "src${pathSeparator}dist",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	// Get expected path separator value (should be OS-specific)
	expectedPathSeparator := string(filepath.Separator)
	
	expectedCommand := "echo " + expectedPathSeparator
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{"path1" + expectedPathSeparator + "path2", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := workspaceDir + expectedPathSeparator + "build"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["PATH_SEP"] != expectedPathSeparator {
		t.Errorf("expected PATH_SEP to be '%s', got '%s'", expectedPathSeparator, substituted.Options.Env["PATH_SEP"])
	}
	
	expectedBuildPath := "src" + expectedPathSeparator + "dist"
	if substituted.Options.Env["BUILD_PATH"] != expectedBuildPath {
		t.Errorf("expected BUILD_PATH to be '%s', got '%s'", expectedBuildPath, substituted.Options.Env["BUILD_PATH"])
	}
}

func TestSubstituteVariables_WithEnvVars(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/main.go"
	
	// Set test environment variables
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("BUILD_TYPE", "debug")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("BUILD_TYPE")
	}()
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${env:TEST_VAR}",
		Args:    []string{"--mode", "${env:BUILD_TYPE}", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}/${env:BUILD_TYPE}",
			Env: map[string]string{
				"CURRENT_VAR":    "${env:TEST_VAR}",
				"BUILD_CONFIG":   "${env:BUILD_TYPE}",
				"MISSING_VAR":    "${env:NONEXISTENT}",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	expectedCommand := "echo test_value"
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{"--mode", "debug", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := workspaceDir + "/debug"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["CURRENT_VAR"] != "test_value" {
		t.Errorf("expected CURRENT_VAR to be 'test_value', got '%s'", substituted.Options.Env["CURRENT_VAR"])
	}
	
	if substituted.Options.Env["BUILD_CONFIG"] != "debug" {
		t.Errorf("expected BUILD_CONFIG to be 'debug', got '%s'", substituted.Options.Env["BUILD_CONFIG"])
	}
	
	// Test that non-existent environment variables are replaced with empty string
	if substituted.Options.Env["MISSING_VAR"] != "" {
		t.Errorf("expected MISSING_VAR to be empty string, got '%s'", substituted.Options.Env["MISSING_VAR"])
	}
}

func TestSubstituteEnvVariables(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_HOME", "/home/test")
	os.Setenv("TEST_PATH", "/usr/bin")
	defer func() {
		os.Unsetenv("TEST_HOME")
		os.Unsetenv("TEST_PATH")
	}()
	
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		{
			input:    "echo ${env:TEST_HOME}",
			expected: "echo /home/test",
			name:     "single env var",
		},
		{
			input:    "${env:TEST_HOME}/${env:TEST_PATH}",
			expected: "/home/test//usr/bin",
			name:     "multiple env vars",
		},
		{
			input:    "no variables here",
			expected: "no variables here",
			name:     "no env vars",
		},
		{
			input:    "${env:NONEXISTENT}",
			expected: "",
			name:     "non-existent env var",
		},
		{
			input:    "prefix_${env:TEST_HOME}_suffix",
			expected: "prefix_/home/test_suffix",
			name:     "env var with prefix and suffix",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := substituteEnvVariables(test.input)
			if result != test.expected {
				t.Errorf("expected '%s', got '%s'", test.expected, result)
			}
		})
	}
}