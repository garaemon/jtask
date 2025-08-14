package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

func TestShouldHandleEvent(t *testing.T) {
	tests := []struct {
		name       string
		event      fsnotify.Event
		extensions []string
		excludes   []string
		expected   bool
	}{
		{
			name:     "write event should be handled",
			event:    fsnotify.Event{Name: "test.go", Op: fsnotify.Write},
			expected: true,
		},
		{
			name:     "create event should be handled",
			event:    fsnotify.Event{Name: "test.go", Op: fsnotify.Create},
			expected: true,
		},
		{
			name:     "remove event should not be handled",
			event:    fsnotify.Event{Name: "test.go", Op: fsnotify.Remove},
			expected: false,
		},
		{
			name:       "file with allowed extension should be handled",
			event:      fsnotify.Event{Name: "test.go", Op: fsnotify.Write},
			extensions: []string{".go", ".js"},
			expected:   true,
		},
		{
			name:       "file with disallowed extension should not be handled",
			event:      fsnotify.Event{Name: "test.txt", Op: fsnotify.Write},
			extensions: []string{".go", ".js"},
			expected:   false,
		},
		{
			name:     "excluded file should not be handled",
			event:    fsnotify.Event{Name: "node_modules/test.js", Op: fsnotify.Write},
			excludes: []string{"node_modules"},
			expected: false,
		},
		{
			name:     "non-excluded file should be handled",
			event:    fsnotify.Event{Name: "src/test.js", Op: fsnotify.Write},
			excludes: []string{"node_modules"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origExtensions := watchExtensions
			origExcludes := watchExclude

			// Set test values
			watchExtensions = tt.extensions
			watchExclude = tt.excludes

			// Test
			result := shouldHandleEvent(tt.event)

			// Restore original values
			watchExtensions = origExtensions
			watchExclude = origExcludes

			if result != tt.expected {
				t.Errorf("shouldHandleEvent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAddWatchPath(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "watch-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create subdirectories
	srcDir := filepath.Join(tempDir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	excludeDir := filepath.Join(tempDir, "node_modules")
	if err := os.Mkdir(excludeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFile := filepath.Join(srcDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	// Save original exclude list
	origExcludes := watchExclude
	watchExclude = []string{"node_modules"}
	defer func() { watchExclude = origExcludes }()

	// Test adding watch path
	err = addWatchPath(watcher, tempDir)
	if err != nil {
		t.Errorf("addWatchPath() error = %v", err)
	}

	// Verify directories are being watched
	// Note: This is a basic test since fsnotify's internal state is not easily inspectable
}

func TestAddWatchPath_NonExistentPath(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	err = addWatchPath(watcher, "/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestExecuteWatchCommand_TaskNotFound(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	
	cmd := &cobra.Command{}
	args := []string{"nonexistent"}
	
	err := executeWatchCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for non-existent task")
	}
	
	if err.Error() != "task 'nonexistent' not found" {
		t.Errorf("expected error message about task not found, got %s", err.Error())
	}
}

func TestExecuteWatchCommand_InvalidTasksFile(t *testing.T) {
	configPath = "../testdata/invalid_tasks.json"
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeWatchCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for invalid tasks file")
	}
}

func TestExecuteWatchCommand_TasksFileNotFound(t *testing.T) {
	configPath = "nonexistent.json"
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeWatchCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for non-existent tasks file")
	}
}

func TestExecuteWatchCommand_WithWorkspaceFolder(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watch-workspace-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create testdata directory and tasks file
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

	// Save original values
	origConfigPath := configPath
	origWorkspaceFolder := workspaceFolder
	origWatchPaths := watchPaths

	// Set test values
	configPath = filepath.Join(testdataDir, "simple_tasks.json")
	workspaceFolder = tempDir
	watchPaths = []string{"/nonexistent"} // This should cause an error

	// Restore original values
	defer func() {
		configPath = origConfigPath
		workspaceFolder = origWorkspaceFolder
		watchPaths = origWatchPaths
	}()

	cmd := &cobra.Command{}
	args := []string{"test-task"}

	// This should fail because we're trying to watch a non-existent path
	err = executeWatchCommand(cmd, args)
	if err == nil {
		t.Error("expected error for non-existent watch path")
	}

	if !strings.Contains(err.Error(), "failed to watch path") {
		t.Errorf("expected watch path error, got %s", err.Error())
	}
}

func TestExecuteWatchCommand_GitRootFallback(t *testing.T) {
	// Create a temporary directory structure with git
	tempDir, err := os.MkdirTemp("", "watch-git-test")
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

	// Save original values
	origConfigPath := configPath
	origWorkspaceFolder := workspaceFolder
	origWatchPaths := watchPaths

	// Set test values
	configPath = "testdata/simple_tasks.json"
	workspaceFolder = ""
	watchPaths = []string{"/nonexistent"} // This should cause an error

	// Restore original values
	defer func() {
		configPath = origConfigPath
		workspaceFolder = origWorkspaceFolder
		watchPaths = origWatchPaths
	}()

	cmd := &cobra.Command{}
	args := []string{"test-task"}

	// This should fail because we're trying to watch a non-existent path
	err = executeWatchCommand(cmd, args)
	if err == nil {
		t.Error("expected error for non-existent watch path")
	}

	if !strings.Contains(err.Error(), "failed to watch path") {
		t.Errorf("expected watch path error, got %s", err.Error())
	}
}

func TestExecuteWatchCommand_NoGitFallbackToCwd(t *testing.T) {
	// Create a temporary directory without git
	tempDir, err := os.MkdirTemp("", "watch-no-git-test")
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

	// Save original values
	origConfigPath := configPath
	origWorkspaceFolder := workspaceFolder
	origWatchPaths := watchPaths

	// Set test values
	configPath = "testdata/simple_tasks.json"
	workspaceFolder = ""
	watchPaths = []string{"/nonexistent"} // This should cause an error

	// Restore original values
	defer func() {
		configPath = origConfigPath
		workspaceFolder = origWorkspaceFolder
		watchPaths = origWatchPaths
	}()

	cmd := &cobra.Command{}
	args := []string{"test-task"}

	// This should fail because we're trying to watch a non-existent path
	err = executeWatchCommand(cmd, args)
	if err == nil {
		t.Error("expected error for non-existent watch path")
	}

	if !strings.Contains(err.Error(), "failed to watch path") {
		t.Errorf("expected watch path error, got %s", err.Error())
	}
}

func TestExecuteWatchCommand_DefaultWatchPaths(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "watch-default-test")
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

	// Save original values
	origConfigPath := configPath
	origWorkspaceFolder := workspaceFolder
	origWatchPaths := watchPaths

	// Set test values - use a non-existent path to force an error and avoid infinite loop
	configPath = filepath.Join(testdataDir, "simple_tasks.json")
	workspaceFolder = "/nonexistent/path"
	watchPaths = []string{} // Empty should default to workspace folder

	// Restore original values
	defer func() {
		configPath = origConfigPath
		workspaceFolder = origWorkspaceFolder
		watchPaths = origWatchPaths
	}()

	cmd := &cobra.Command{}
	args := []string{"test-task"}

	// This should fail because we're trying to watch a non-existent workspace folder
	err = executeWatchCommand(cmd, args)
	if err == nil {
		t.Error("expected error for non-existent workspace folder")
	}

	if !strings.Contains(err.Error(), "failed to watch path") {
		t.Errorf("expected watch path error, got %s", err.Error())
	}
}