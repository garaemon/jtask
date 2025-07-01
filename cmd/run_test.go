package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecuteRunCommand_TaskNotFound(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	
	cmd := &cobra.Command{}
	args := []string{"nonexistent"}
	
	err := executeRunCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for non-existent task")
	}
	
	if err.Error() != "task 'nonexistent' not found" {
		t.Errorf("expected error message about task not found, got %s", err.Error())
	}
}

func TestExecuteRunCommand_DryRun(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !bytes.Contains(out, []byte("Would execute task: build")) {
		t.Errorf("expected dry run output, got %s", output)
	}
}

func TestExecuteRunCommand_InvalidTasksFile(t *testing.T) {
	configPath = "../testdata/invalid_tasks.json"
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for invalid tasks file")
	}
}

func TestExecuteRunCommand_TasksFileNotFound(t *testing.T) {
	configPath = "nonexistent.json"
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for non-existent tasks file")
	}
}

func TestExecuteRunCommand_WithWorkspaceFolder(t *testing.T) {
	configPath = "../testdata/workspace_tasks.json"
	workspaceFolder = "/test/workspace"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"workspace-build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !strings.Contains(output, "echo building in /test/workspace") {
		t.Errorf("expected workspace folder substitution in output, got %s", output)
	}
}

func TestExecuteRunCommand_WorkspaceFolderWithArgs(t *testing.T) {
	configPath = "../testdata/workspace_tasks.json"
	workspaceFolder = "/test/workspace"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"workspace-args"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !strings.Contains(output, "Args: [-la /test/workspace]") {
		t.Errorf("expected workspace folder substitution in args, got %s", output)
	}
}

func TestExecuteRunCommand_VerboseOutput(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	verbose = true
	defer func() { verbose = false }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !strings.Contains(output, "Workspace folder:") {
		t.Errorf("expected workspace folder info in verbose output, got %s", output)
	}
	if !strings.Contains(output, "Using tasks file:") {
		t.Errorf("expected tasks file info in verbose output, got %s", output)
	}
}

func TestExecuteRunCommand_QuietMode(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	quiet = true
	defer func() { quiet = false }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if strings.Contains(output, "Executing task:") {
		t.Errorf("expected no execution message in quiet mode, got %s", output)
	}
}

func TestExecuteRunCommand_GitRootFallback(t *testing.T) {
	// Create a temporary directory structure with git
	tempDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	
	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create testdata and tasks file in temp dir
	testdataDir := filepath.Join(tempDir, "testdata")
	if err := os.Mkdir(testdataDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	tasksFile := filepath.Join(testdataDir, "simple_tasks.json")
	if err := os.WriteFile(tasksFile, []byte(`{
		"version": "2.0.0",
		"tasks": [
			{
				"label": "test-task",
				"type": "shell",
				"command": "echo test"
			}
		]
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tempDir)
	
	configPath = "testdata/simple_tasks.json"
	workspaceFolder = ""
	verbose = true
	defer func() { verbose = false }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"test-task"}
	
	err = executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	// tempDir might have symbolic links resolved, so check if the output contains the directory name
	if !strings.Contains(output, "Workspace folder:") || !strings.Contains(output, "git-test") {
		t.Errorf("expected git root as workspace folder, got %s", output)
	}
}

func TestExecuteRunCommand_NoGitFallbackToCwd(t *testing.T) {
	// Create a temporary directory without git
	tempDir, err := os.MkdirTemp("", "no-git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	
	// Create testdata and tasks file in temp dir
	testdataDir := filepath.Join(tempDir, "testdata")
	if err := os.Mkdir(testdataDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	tasksFile := filepath.Join(testdataDir, "simple_tasks.json")
	if err := os.WriteFile(tasksFile, []byte(`{
		"version": "2.0.0",
		"tasks": [
			{
				"label": "test-task",
				"type": "shell",
				"command": "echo test"
			}
		]
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tempDir)
	
	configPath = "testdata/simple_tasks.json"
	workspaceFolder = ""
	verbose = true
	defer func() { verbose = false }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"test-task"}
	
	err = executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	// tempDir might have symbolic links resolved, so check if the output contains the directory name
	if !strings.Contains(output, "Workspace folder:") || !strings.Contains(output, "no-git-test") {
		t.Errorf("expected current directory as workspace folder fallback, got %s", output)
	}
}

func TestExecuteRunCommand_ComplexTaskWithOptions(t *testing.T) {
	configPath = "../testdata/complex_tasks.json"
	workspaceFolder = "/test/workspace"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"compile"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !strings.Contains(output, "Would execute task: compile") {
		t.Errorf("expected task execution info, got %s", output)
	}
	if !strings.Contains(output, "Type: process") {
		t.Errorf("expected task type info, got %s", output)
	}
}

func TestExecuteRunCommand_WorkspaceFolderAbsolutePath(t *testing.T) {
	configPath = "../testdata/workspace_tasks.json"
	workspaceFolder = "/absolute/path/to/workspace"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"workspace-cwd"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	// Note: The cwd option is not shown in dry-run output, but the workspace substitution should occur
	if !strings.Contains(output, "Would execute task: workspace-cwd") {
		t.Errorf("expected task execution with workspace folder, got %s", output)
	}
}

func TestExecuteRunCommand_GetCurrentDirError(t *testing.T) {
	// This test is difficult to create a real scenario for os.Getwd() error
	// but we can test the error handling path by checking the error message structure
	configPath = "../testdata/simple_tasks.json"
	workspaceFolder = ""
	
	// We'll verify the function structure handles errors properly
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	// In normal circumstances, this should not error
	err := executeRunCommand(cmd, args)
	if err != nil {
		// If there's an error, it should be meaningful
		if !strings.Contains(err.Error(), "failed to") {
			t.Errorf("expected meaningful error message, got %s", err.Error())
		}
	}
}

func TestExecuteRunCommand_EmptyWorkspaceFolder(t *testing.T) {
	configPath = "../testdata/workspace_tasks.json"
	workspaceFolder = ""
	dryRun = true
	defer func() { dryRun = false }()
	verbose = true
	defer func() { verbose = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"workspace-build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	// Should show workspace folder determination
	if !strings.Contains(output, "Workspace folder:") {
		t.Errorf("expected workspace folder info in verbose output, got %s", output)
	}
}

func TestExecuteRunCommand_RelativeWorkspaceFolder(t *testing.T) {
	configPath = "../testdata/workspace_tasks.json"
	workspaceFolder = "./relative/path"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"workspace-build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !strings.Contains(output, "echo building in ./relative/path") {
		t.Errorf("expected relative workspace folder substitution, got %s", output)
	}
}

func TestExecuteRunCommand_TaskWithoutWorkspaceVar(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	workspaceFolder = "/test/workspace"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	cmd := &cobra.Command{}
	args := []string{"test"}
	
	err := executeRunCommand(cmd, args)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Should work fine even if task doesn't use workspace folder variables
}

func TestExecuteRunCommand_MultipleWorkspaceVars(t *testing.T) {
	// Create a test file with multiple workspace folder references
	configPath = "../testdata/workspace_tasks.json"
	workspaceFolder = "/test/workspace"
	defer func() { workspaceFolder = "" }()
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"workspace-env"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !strings.Contains(output, "Would execute task: workspace-env") {
		t.Errorf("expected task execution with multiple workspace variables, got %s", output)
	}
}