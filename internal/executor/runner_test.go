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

func TestSubstituteVariables_WithWorkspaceFolderBasename(t *testing.T) {
	workspaceDir := "/home/user/my-project"
	file := "src/main.go"
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${workspaceFolderBasename}",
		Args:    []string{"--project", "${workspaceFolderBasename}", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}/${workspaceFolderBasename}-build",
			Env: map[string]string{
				"PROJECT_NAME": "${workspaceFolderBasename}",
				"BUILD_DIR":    "${workspaceFolderBasename}/dist",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	expectedCommand := "echo my-project"
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{"--project", "my-project", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := workspaceDir + "/my-project-build"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["PROJECT_NAME"] != "my-project" {
		t.Errorf("expected PROJECT_NAME to be 'my-project', got '%s'", substituted.Options.Env["PROJECT_NAME"])
	}
	
	expectedBuildDir := "my-project/dist"
	if substituted.Options.Env["BUILD_DIR"] != expectedBuildDir {
		t.Errorf("expected BUILD_DIR to be '%s', got '%s'", expectedBuildDir, substituted.Options.Env["BUILD_DIR"])
	}
}

func TestSubstituteVariables_WorkspaceFolderBasenameEdgeCases(t *testing.T) {
	tests := []struct {
		workspaceDir string
		expected     string
		name         string
	}{
		{
			workspaceDir: "/home/user/project",
			expected:     "project",
			name:         "normal path",
		},
		{
			workspaceDir: "/home/user/my-project-with-dashes",
			expected:     "my-project-with-dashes",
			name:         "path with dashes",
		},
		{
			workspaceDir: "/",
			expected:     "/",
			name:         "root directory",
		},
		{
			workspaceDir: "project",
			expected:     "project",
			name:         "relative path",
		},
		{
			workspaceDir: "/home/user/project/",
			expected:     "project",
			name:         "path with trailing slash",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := &config.Task{
				Type:    "shell",
				Command: "echo ${workspaceFolderBasename}",
			}
			
			substituted := substituteVariables(task, test.workspaceDir, "")
			expectedCommand := "echo " + test.expected
			
			if substituted.Command != expectedCommand {
				t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
			}
		})
	}
}

func TestSubstituteVariables_WithFileBasename(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/components/main.tsx"
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${fileBasename}",
		Args:    []string{"--file", "${fileBasename}", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}/build",
			Env: map[string]string{
				"CURRENT_FILE": "${fileBasename}",
				"OUTPUT_FILE": "${fileBasename}.bak",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	expectedCommand := "echo main.tsx"
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{"--file", "main.tsx", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := workspaceDir + "/build"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["CURRENT_FILE"] != "main.tsx" {
		t.Errorf("expected CURRENT_FILE to be 'main.tsx', got '%s'", substituted.Options.Env["CURRENT_FILE"])
	}
	
	if substituted.Options.Env["OUTPUT_FILE"] != "main.tsx.bak" {
		t.Errorf("expected OUTPUT_FILE to be 'main.tsx.bak', got '%s'", substituted.Options.Env["OUTPUT_FILE"])
	}
}

func TestSubstituteVariables_FileBasenameEdgeCases(t *testing.T) {
	tests := []struct {
		file     string
		expected string
		name     string
	}{
		{
			file:     "src/main.go",
			expected: "main.go",
			name:     "normal file path",
		},
		{
			file:     "src/components/Button.tsx",
			expected: "Button.tsx",
			name:     "nested file path",
		},
		{
			file:     "main.go",
			expected: "main.go",
			name:     "file in root",
		},
		{
			file:     "README.md",
			expected: "README.md",
			name:     "markdown file",
		},
		{
			file:     "src/utils/helper.test.js",
			expected: "helper.test.js",
			name:     "test file with multiple dots",
		},
		{
			file:     "file-with-dashes.json",
			expected: "file-with-dashes.json",
			name:     "file with dashes",
		},
		{
			file:     "",
			expected: "",
			name:     "empty file path",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := &config.Task{
				Type:    "shell",
				Command: "echo ${fileBasename}",
			}
			
			substituted := substituteVariables(task, "/home/user/project", test.file)
			expectedCommand := "echo " + test.expected
			
			if substituted.Command != expectedCommand {
				t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
			}
		})
	}
}

func TestSubstituteVariables_WithFileBasenameNoExtension(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/components/main.tsx"
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${fileBasenameNoExtension}",
		Args:    []string{"--name", "${fileBasenameNoExtension}", "test"},
		Options: &config.TaskOptions{
			Cwd: "${workspaceFolder}/build",
			Env: map[string]string{
				"FILE_NAME": "${fileBasenameNoExtension}",
				"OUTPUT_FILE": "${fileBasenameNoExtension}.compiled.js",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	expectedCommand := "echo main"
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{"--name", "main", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := workspaceDir + "/build"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["FILE_NAME"] != "main" {
		t.Errorf("expected FILE_NAME to be 'main', got '%s'", substituted.Options.Env["FILE_NAME"])
	}
	
	if substituted.Options.Env["OUTPUT_FILE"] != "main.compiled.js" {
		t.Errorf("expected OUTPUT_FILE to be 'main.compiled.js', got '%s'", substituted.Options.Env["OUTPUT_FILE"])
	}
}

func TestSubstituteVariables_FileBasenameNoExtensionEdgeCases(t *testing.T) {
	tests := []struct {
		file     string
		expected string
		name     string
	}{
		{
			file:     "src/main.go",
			expected: "main",
			name:     "normal file with extension",
		},
		{
			file:     "src/components/Button.tsx",
			expected: "Button",
			name:     "nested file with extension",
		},
		{
			file:     "main.go",
			expected: "main",
			name:     "file in root with extension",
		},
		{
			file:     "README.md",
			expected: "README",
			name:     "markdown file",
		},
		{
			file:     "src/utils/helper.test.js",
			expected: "helper.test",
			name:     "test file with multiple dots",
		},
		{
			file:     "file-with-dashes.json",
			expected: "file-with-dashes",
			name:     "file with dashes",
		},
		{
			file:     "noextension",
			expected: "noextension",
			name:     "file without extension",
		},
		{
			file:     "src/utils/.gitignore",
			expected: "",
			name:     "dotfile with extension",
		},
		{
			file:     "",
			expected: "",
			name:     "empty file path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := &config.Task{
				Type:    "shell",
				Command: "echo ${fileBasenameNoExtension}",
			}
			
			substituted := substituteVariables(task, "/home/user/project", test.file)
			expectedCommand := "echo " + test.expected
			
			if substituted.Command != expectedCommand {
				t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
			}
		})
	}
}

func TestSubstituteVariables_WithFileDirname(t *testing.T) {
	workspaceDir := "/home/user/project"
	file := "src/components/main.tsx"
	
	task := &config.Task{
		Type:    "shell",
		Command: "echo ${fileDirname}",
		Args:    []string{"--dir", "${fileDirname}", "test"},
		Options: &config.TaskOptions{
			Cwd: "${fileDirname}/build",
			Env: map[string]string{
				"SOURCE_DIR": "${fileDirname}",
				"BACKUP_DIR": "${fileDirname}/.backup",
			},
		},
	}
	
	substituted := substituteVariables(task, workspaceDir, file)
	
	expectedCommand := "echo src/components"
	if substituted.Command != expectedCommand {
		t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
	}
	
	expectedArgs := []string{"--dir", "src/components", "test"}
	if len(substituted.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(substituted.Args))
	}
	
	for i, arg := range expectedArgs {
		if substituted.Args[i] != arg {
			t.Errorf("expected arg %d to be %s, got %s", i, arg, substituted.Args[i])
		}
	}
	
	expectedCwdPath := "src/components/build"
	if substituted.Options.Cwd != expectedCwdPath {
		t.Errorf("expected cwd to be '%s', got '%s'", expectedCwdPath, substituted.Options.Cwd)
	}
	
	if substituted.Options.Env["SOURCE_DIR"] != "src/components" {
		t.Errorf("expected SOURCE_DIR to be 'src/components', got '%s'", substituted.Options.Env["SOURCE_DIR"])
	}
	
	if substituted.Options.Env["BACKUP_DIR"] != "src/components/.backup" {
		t.Errorf("expected BACKUP_DIR to be 'src/components/.backup', got '%s'", substituted.Options.Env["BACKUP_DIR"])
	}
}

func TestSubstituteVariables_FileDirnameEdgeCases(t *testing.T) {
	tests := []struct {
		file     string
		expected string
		name     string
	}{
		{
			file:     "src/main.go",
			expected: "src",
			name:     "normal file path",
		},
		{
			file:     "src/components/Button.tsx",
			expected: "src/components",
			name:     "nested file path",
		},
		{
			file:     "main.go",
			expected: ".",
			name:     "file in root",
		},
		{
			file:     "deep/nested/folder/file.js",
			expected: "deep/nested/folder",
			name:     "deeply nested file",
		},
		{
			file:     "/absolute/path/file.txt",
			expected: "/absolute/path",
			name:     "absolute path",
		},
		{
			file:     "folder/subfolder/",
			expected: "folder/subfolder",
			name:     "trailing slash",
		},
		{
			file:     "",
			expected: "",
			name:     "empty file path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			task := &config.Task{
				Type:    "shell",
				Command: "echo ${fileDirname}",
			}
			
			substituted := substituteVariables(task, "/home/user/project", test.file)
			expectedCommand := "echo " + test.expected
			
			if substituted.Command != expectedCommand {
				t.Errorf("expected command to be '%s', got '%s'", expectedCommand, substituted.Command)
			}
		})
	}
}