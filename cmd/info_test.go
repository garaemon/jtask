package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/garaemon/tasks-json-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestRunInfoCommand(t *testing.T) {
	// Save original values
	origVerbose := verbose
	origQuiet := quiet
	origConfigPath := configPath
	defer func() {
		verbose = origVerbose
		quiet = origQuiet
		configPath = origConfigPath
	}()

	tests := []struct {
		name           string
		taskName       string
		configPath     string
		expectError    bool
		expectedOutput string
	}{
		{
			name:       "existing task",
			taskName:   "build",
			configPath: "../testdata/simple_tasks.json",
			expectError: false,
			expectedOutput: "Task: build",
		},
		{
			name:       "non-existent task",
			taskName:   "nonexistent",
			configPath: "../testdata/simple_tasks.json",
			expectError: true,
		},
		{
			name:       "invalid config path",
			taskName:   "build",
			configPath: "nonexistent.json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			verbose = false
			quiet = false
			configPath = tt.configPath

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run command
			cmd := &cobra.Command{}
			err := runInfoCommand(cmd, []string{tt.taskName})

			// Restore stdout
			_ = w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("expected output to contain '%s', got: %s", tt.expectedOutput, output)
				}
			}
		})
	}
}

func TestFindTaskByName(t *testing.T) {
	tasks := []config.Task{
		{Label: "build", Type: "shell", Command: "go build"},
		{Label: "test", Type: "shell", Command: "go test"},
	}

	tests := []struct {
		name     string
		taskName string
		expected *config.Task
	}{
		{
			name:     "existing task",
			taskName: "build",
			expected: &tasks[0],
		},
		{
			name:     "non-existent task",
			taskName: "nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findTaskByName(tasks, tt.taskName)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected task, got nil")
				} else if result.Label != tt.expected.Label {
					t.Errorf("expected label %s, got %s", tt.expected.Label, result.Label)
				}
			}
		})
	}
}

func TestPrintTaskInfo(t *testing.T) {
	// Save original values
	origVerbose := verbose
	origQuiet := quiet
	defer func() {
		verbose = origVerbose
		quiet = origQuiet
	}()

	task := &config.Task{
		Label:   "test-task",
		Type:    "shell",
		Command: "echo hello",
		Args:    []string{"world"},
		Group:   "build",
		Options: &config.TaskOptions{
			Cwd: "/tmp",
			Env: map[string]string{
				"TEST_VAR": "test_value",
			},
		},
		DependsOn: []string{"other-task"},
	}

	tests := []struct {
		name           string
		verbose        bool
		quiet          bool
		expectedOutput []string
	}{
		{
			name:    "normal output",
			verbose: false,
			quiet:   false,
			expectedOutput: []string{
				"Task: test-task",
				"Type:     shell",
				"Command:  echo hello",
				"Args:     world",
				"Group:    build",
				"Working Directory: /tmp",
				"TEST_VAR=test_value",
				"Depends On: other-task",
			},
		},
		{
			name:    "verbose output",
			verbose: true,
			quiet:   false,
			expectedOutput: []string{
				"Task: test-task",
				"Additional Information:",
				"Source File:",
			},
		},
		{
			name:    "quiet output",
			verbose: false,
			quiet:   true,
			expectedOutput: []string{
				"test-task\tshell\techo hello world",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			verbose = tt.verbose
			quiet = tt.quiet

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run function
			printTaskInfo(task, "/path/to/tasks.json")

			// Restore stdout
			_ = w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			// Check all expected strings are present
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}