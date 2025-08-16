package cmd

import (
	"encoding/json"
	"os"
	"testing"
)

func TestValidateTasksFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectValid bool
		expectErrors int
		expectWarnings int
	}{
		{
			name: "valid tasks.json",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build",
						"type": "shell",
						"command": "go build"
					}
				]
			}`,
			expectValid: true,
			expectErrors: 0,
			expectWarnings: 0,
		},
		{
			name: "invalid JSON syntax",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build",
						"type": "shell",
						"command": "go build"
					},
				]
			}`,
			expectValid: false,
			expectErrors: 1,
			expectWarnings: 0,
		},
		{
			name: "npm task valid",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "start",
						"type": "npm",
						"script": "start"
					}
				]
			}`,
			expectValid: true,
			expectErrors: 0,
			expectWarnings: 0,
		},
		{
			name: "typescript task valid",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build",
						"type": "typescript",
						"tsconfig": "tsconfig.json"
					}
				]
			}`,
			expectValid: true,
			expectErrors: 0,
			expectWarnings: 0,
		},
		{
			name: "npm task missing script",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "start",
						"type": "npm"
					}
				]
			}`,
			expectValid: false,
			expectErrors: 1,
			expectWarnings: 0,
		},
		{
			name: "missing required fields",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build"
					}
				]
			}`,
			expectValid: false,
			expectErrors: 1,
			expectWarnings: 1,
		},
		{
			name: "duplicate labels",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build",
						"type": "shell",
						"command": "go build"
					},
					{
						"label": "build",
						"type": "shell",
						"command": "go test"
					}
				]
			}`,
			expectValid: false,
			expectErrors: 1,
			expectWarnings: 0,
		},
		{
			name: "unknown task type",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build",
						"type": "unknown",
						"command": "go build"
					}
				]
			}`,
			expectValid: true,
			expectErrors: 0,
			expectWarnings: 1,
		},
		{
			name: "unknown dependency",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "build",
						"type": "shell",
						"command": "go build",
						"dependsOn": "nonexistent"
					}
				]
			}`,
			expectValid: true,
			expectErrors: 0,
			expectWarnings: 1,
		},
		{
			name: "valid dependency",
			content: `{
				"version": "2.0.0",
				"tasks": [
					{
						"label": "deps",
						"type": "shell",
						"command": "go mod download"
					},
					{
						"label": "build",
						"type": "shell",
						"command": "go build",
						"dependsOn": "deps"
					}
				]
			}`,
			expectValid: true,
			expectErrors: 0,
			expectWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "tasks_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			// Write test content
			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			_ = tmpFile.Close()

			// Validate the file
			result := validateTasksFile(tmpFile.Name())

			// Check results
			if result.Valid != tt.expectValid {
				resultData, _ := json.MarshalIndent(result, "", "  ")
				t.Errorf("expected valid=%v, got valid=%v\nResult: %s", 
					tt.expectValid, result.Valid, string(resultData))
			}

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("expected %d errors, got %d errors: %v", 
					tt.expectErrors, len(result.Errors), result.Errors)
			}

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("expected %d warnings, got %d warnings: %v", 
					tt.expectWarnings, len(result.Warnings), result.Warnings)
			}
		})
	}
}

func TestValidateNonExistentFile(t *testing.T) {
	result := validateTasksFile("/nonexistent/path/tasks.json")
	
	if result.Valid {
		t.Error("expected invalid result for non-existent file")
	}
	
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	
	if result.Errors[0].Type != "file_not_found" {
		t.Errorf("expected file_not_found error, got %s", result.Errors[0].Type)
	}
}

func TestValidateWorkingDirectory(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "validate_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create tasks.json with valid working directory
	validContent := `{
		"version": "2.0.0",
		"tasks": [
			{
				"label": "build",
				"type": "shell",
				"command": "go build",
				"options": {
					"cwd": "` + tmpDir + `"
				}
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "tasks_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(validContent); err != nil {
		t.Fatal(err)
	}
	_ = tmpFile.Close()

	result := validateTasksFile(tmpFile.Name())
	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings for valid directory, got: %v", result.Warnings)
	}

	// Test with invalid working directory
	invalidContent := `{
		"version": "2.0.0",
		"tasks": [
			{
				"label": "build",
				"type": "shell",
				"command": "go build",
				"options": {
					"cwd": "/nonexistent/directory"
				}
			}
		]
	}`

	tmpFile2, err := os.CreateTemp("", "tasks_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile2.Name()) }()

	if _, err := tmpFile2.WriteString(invalidContent); err != nil {
		t.Fatal(err)
	}
	_ = tmpFile2.Close()

	result2 := validateTasksFile(tmpFile2.Name())
	if !result2.Valid {
		t.Errorf("expected valid result (should only warn), got errors: %v", result2.Errors)
	}

	if len(result2.Warnings) != 1 {
		t.Errorf("expected 1 warning for invalid directory, got: %v", result2.Warnings)
	}
}

func TestGetDependsOnAsStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "string input",
			input:    "build",
			expected: []string{"build"},
		},
		{
			name:     "string slice input",
			input:    []string{"deps", "clean"},
			expected: []string{"deps", "clean"},
		},
		{
			name:     "interface slice input",
			input:    []interface{}{"deps", "clean"},
			expected: []string{"deps", "clean"},
		},
		{
			name:     "mixed interface slice",
			input:    []interface{}{"deps", 123, "clean"},
			expected: []string{"deps", "clean"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDependsOnAsStringSlice(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("expected %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}