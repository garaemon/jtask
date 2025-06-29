package config

import (
	"path/filepath"
	"testing"
)

func TestLoadTasks_SimpleFile(t *testing.T) {
	testFile := filepath.Join("..", "..", "testdata", "simple_tasks.json")
	
	tasks, err := LoadTasks(testFile)
	if err != nil {
		t.Fatalf("LoadTasks failed: %v", err)
	}
	
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
	
	buildTask := tasks[0]
	if buildTask.Label != "build" {
		t.Errorf("Expected label 'build', got '%s'", buildTask.Label)
	}
	if buildTask.Type != "shell" {
		t.Errorf("Expected type 'shell', got '%s'", buildTask.Type)
	}
	if buildTask.Command != "go build" {
		t.Errorf("Expected command 'go build', got '%s'", buildTask.Command)
	}
	if buildTask.GetGroupKind() != "build" {
		t.Errorf("Expected group 'build', got '%s'", buildTask.GetGroupKind())
	}
	
	testTask := tasks[1]
	if testTask.Label != "test" {
		t.Errorf("Expected label 'test', got '%s'", testTask.Label)
	}
	if len(testTask.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(testTask.Args))
	}
	if testTask.Args[0] != "-v" || testTask.Args[1] != "./..." {
		t.Errorf("Expected args ['-v', './...'], got %v", testTask.Args)
	}
}

func TestLoadTasks_ComplexFile(t *testing.T) {
	testFile := filepath.Join("..", "..", "testdata", "complex_tasks.json")
	
	tasks, err := LoadTasks(testFile)
	if err != nil {
		t.Fatalf("LoadTasks failed: %v", err)
	}
	
	if len(tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(tasks))
	}
	
	compileTask := tasks[0]
	if compileTask.Label != "compile" {
		t.Errorf("Expected label 'compile', got '%s'", compileTask.Label)
	}
	if compileTask.Type != "process" {
		t.Errorf("Expected type 'process', got '%s'", compileTask.Type)
	}
	if compileTask.GetGroupKind() != "build" {
		t.Errorf("Expected group 'build', got '%s'", compileTask.GetGroupKind())
	}
	if !compileTask.IsDefaultInGroup() {
		t.Error("Expected compile task to be default in group")
	}
	
	if compileTask.Options == nil {
		t.Error("Expected options to be set")
	} else {
		if compileTask.Options.Cwd != "${workspaceFolder}/src" {
			t.Errorf("Expected cwd '${workspaceFolder}/src', got '%s'", compileTask.Options.Cwd)
		}
		if compileTask.Options.Env["DEBUG"] != "1" {
			t.Errorf("Expected DEBUG env var '1', got '%s'", compileTask.Options.Env["DEBUG"])
		}
	}
	
	watchTask := tasks[2]
	if watchTask.Label != "watch" {
		t.Errorf("Expected label 'watch', got '%s'", watchTask.Label)
	}
	if watchTask.GetGroupKind() != "" {
		t.Errorf("Expected no group, got '%s'", watchTask.GetGroupKind())
	}
	if watchTask.IsDefaultInGroup() {
		t.Error("Expected watch task not to be default in group")
	}
}

func TestLoadTasks_FileNotFound(t *testing.T) {
	_, err := LoadTasks("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestLoadTasks_InvalidJSON(t *testing.T) {
	testFile := filepath.Join("..", "..", "testdata", "invalid_tasks.json")
	
	_, err := LoadTasks(testFile)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestTask_GetGroupKind(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		expected string
	}{
		{
			name:     "nil group",
			task:     Task{Group: nil},
			expected: "",
		},
		{
			name:     "string group",
			task:     Task{Group: "build"},
			expected: "build",
		},
		{
			name: "object group with kind",
			task: Task{Group: map[string]interface{}{
				"kind":      "test",
				"isDefault": true,
			}},
			expected: "test",
		},
		{
			name: "object group without kind",
			task: Task{Group: map[string]interface{}{
				"isDefault": true,
			}},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.GetGroupKind()
			if result != tt.expected {
				t.Errorf("GetGroupKind() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestTask_IsDefaultInGroup(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		expected bool
	}{
		{
			name:     "nil group",
			task:     Task{Group: nil},
			expected: false,
		},
		{
			name:     "string group",
			task:     Task{Group: "build"},
			expected: false,
		},
		{
			name: "object group with isDefault true",
			task: Task{Group: map[string]interface{}{
				"kind":      "build",
				"isDefault": true,
			}},
			expected: true,
		},
		{
			name: "object group with isDefault false",
			task: Task{Group: map[string]interface{}{
				"kind":      "build",
				"isDefault": false,
			}},
			expected: false,
		},
		{
			name: "object group without isDefault",
			task: Task{Group: map[string]interface{}{
				"kind": "build",
			}},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.IsDefaultInGroup()
			if result != tt.expected {
				t.Errorf("IsDefaultInGroup() = %v, expected %v", result, tt.expected)
			}
		})
	}
}