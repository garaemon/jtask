package executor

import (
	"testing"
	"path/filepath"
	"strings"

	"github.com/garaemon/tasks-json-cli/internal/config"
)

func TestBuildNpmCommand(t *testing.T) {
	workspaceDir := "/workspace"
	
	tests := []struct {
		name         string
		task         *config.Task
		expectedCmd  string
		expectedArgs []string
		expectedDir  string
		expectError  bool
	}{
		{
			name: "simple npm task",
			task: &config.Task{
				Type:   "npm",
				Script: "start",
			},
			expectedCmd:  "npm",
			expectedArgs: []string{"run", "start"},
			expectedDir:  workspaceDir,
			expectError:  false,
		},
		{
			name: "npm task with path",
			task: &config.Task{
				Type:   "npm",
				Script: "build",
				Path:   "frontend",
			},
			expectedCmd:  "npm",
			expectedArgs: []string{"run", "build"},
			expectedDir:  filepath.Join(workspaceDir, "frontend"),
			expectError:  false,
		},
		{
			name: "npm task with absolute path",
			task: &config.Task{
				Type:   "npm",
				Script: "test",
				Path:   "/absolute/path",
			},
			expectedCmd:  "npm",
			expectedArgs: []string{"run", "test"},
			expectedDir:  "/absolute/path",
			expectError:  false,
		},
		{
			name: "npm task without script",
			task: &config.Task{
				Type: "npm",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := buildNpmCommand(tt.task, workspaceDir)
			
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			// Check that the command ends with npm (path may vary)
			if !strings.HasSuffix(cmd.Path, "npm") && cmd.Path != tt.expectedCmd {
				t.Errorf("expected command to end with %q, got %q", tt.expectedCmd, cmd.Path)
			}
			
			if len(cmd.Args) != len(tt.expectedArgs)+1 { // +1 for command name
				t.Errorf("expected %d args, got %d", len(tt.expectedArgs), len(cmd.Args)-1)
			} else {
				for i, expectedArg := range tt.expectedArgs {
					if cmd.Args[i+1] != expectedArg { // +1 to skip command name
						t.Errorf("expected arg[%d] %q, got %q", i, expectedArg, cmd.Args[i+1])
					}
				}
			}
			
			if cmd.Dir != tt.expectedDir {
				t.Errorf("expected dir %q, got %q", tt.expectedDir, cmd.Dir)
			}
		})
	}
}