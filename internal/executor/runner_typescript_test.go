package executor

import (
	"testing"
	"strings"

	"github.com/garaemon/tasks-json-cli/internal/config"
)

func TestBuildTypescriptCommand(t *testing.T) {
	workspaceDir := "/workspace"
	
	tests := []struct {
		name         string
		task         *config.Task
		expectedCmd  string
		expectedArgs []string
		expectedDir  string
	}{
		{
			name: "simple typescript task",
			task: &config.Task{
				Type: "typescript",
			},
			expectedCmd:  "tsc",
			expectedArgs: []string{},
			expectedDir:  workspaceDir,
		},
		{
			name: "typescript task with tsconfig",
			task: &config.Task{
				Type:     "typescript",
				TSConfig: "tsconfig.json",
			},
			expectedCmd:  "tsc",
			expectedArgs: []string{"-p", "tsconfig.json"},
			expectedDir:  workspaceDir,
		},
		{
			name: "typescript task with watch option",
			task: &config.Task{
				Type:   "typescript",
				Option: "watch",
			},
			expectedCmd:  "tsc",
			expectedArgs: []string{"--watch"},
			expectedDir:  workspaceDir,
		},
		{
			name: "typescript task with tsconfig and watch",
			task: &config.Task{
				Type:     "typescript",
				TSConfig: "tsconfig.build.json",
				Option:   "watch",
			},
			expectedCmd:  "tsc",
			expectedArgs: []string{"-p", "tsconfig.build.json", "--watch"},
			expectedDir:  workspaceDir,
		},
		{
			name: "typescript task with custom option",
			task: &config.Task{
				Type:   "typescript",
				Option: "--noEmit",
			},
			expectedCmd:  "tsc",
			expectedArgs: []string{"--noEmit"},
			expectedDir:  workspaceDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := buildTypescriptCommand(tt.task, workspaceDir)
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			// Check that the command ends with tsc (path may vary)
			if !strings.HasSuffix(cmd.Path, "tsc") && cmd.Path != tt.expectedCmd {
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