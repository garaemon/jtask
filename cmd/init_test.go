package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCommand(t *testing.T) {
	// Save original values
	origVerbose := verbose
	origQuiet := quiet
	origTemplateName := templateName
	origForceFlag := forceFlag
	origOutputPath := outputPath
	defer func() {
		verbose = origVerbose
		quiet = origQuiet
		templateName = origTemplateName
		forceFlag = origForceFlag
		outputPath = origOutputPath
	}()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "tasks-json-cli-init-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name         string
		template     string
		outputPath   string
		forceFlag    bool
		setupFunc    func() error
		expectError  bool
		checkContent bool
	}{
		{
			name:         "default template",
			template:     "default",
			outputPath:   filepath.Join(tempDir, "tasks1.json"),
			forceFlag:    false,
			expectError:  false,
			checkContent: true,
		},
		{
			name:         "go template",
			template:     "go",
			outputPath:   filepath.Join(tempDir, "tasks2.json"),
			forceFlag:    false,
			expectError:  false,
			checkContent: true,
		},
		{
			name:         "node template",
			template:     "node",
			outputPath:   filepath.Join(tempDir, "tasks3.json"),
			forceFlag:    false,
			expectError:  false,
			checkContent: true,
		},
		{
			name:        "invalid template",
			template:    "invalid",
			outputPath:  filepath.Join(tempDir, "tasks4.json"),
			forceFlag:   false,
			expectError: true,
		},
		{
			name:       "force overwrite existing file",
			template:   "default",
			outputPath: filepath.Join(tempDir, "tasks5.json"),
			forceFlag:  true,
			setupFunc: func() error {
				return os.WriteFile(filepath.Join(tempDir, "tasks5.json"), []byte("existing content"), 0644)
			},
			expectError:  false,
			checkContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			verbose = false
			quiet = true
			templateName = tt.template
			forceFlag = tt.forceFlag
			outputPath = tt.outputPath

			// Setup function if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Run command
			err := runInitCommand(nil, []string{})

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check if file was created
				if _, err := os.Stat(tt.outputPath); os.IsNotExist(err) {
					t.Errorf("expected file to be created at %s", tt.outputPath)
				}

				// Check content if requested
				if tt.checkContent {
					content, err := os.ReadFile(tt.outputPath)
					if err != nil {
						t.Errorf("failed to read created file: %v", err)
					} else {
						contentStr := string(content)
						if !strings.Contains(contentStr, `"version": "2.0.0"`) {
							t.Errorf("expected tasks.json format, got: %s", contentStr)
						}
					}
				}
			}
		})
	}
}

func TestIsValidTemplate(t *testing.T) {
	tests := []struct {
		template string
		expected bool
	}{
		{"default", true},
		{"go", true},
		{"node", true},
		{"invalid", false},
		{"", false},
		{"Default", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			result := isValidTemplate(tt.template)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateDirectoryIfNeeded(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tasks-json-cli-dir-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "create nested directory",
			filePath:    filepath.Join(tempDir, "nested", "dir", "file.json"),
			expectError: false,
		},
		{
			name:        "file in current directory",
			filePath:    "file.json",
			expectError: false,
		},
		{
			name:        "existing directory",
			filePath:    filepath.Join(tempDir, "file.json"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createDirectoryIfNeeded(tt.filePath)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check if directory was created (except for current directory case)
				if tt.filePath != "file.json" {
					dir := filepath.Dir(tt.filePath)
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						t.Errorf("expected directory to be created: %s", dir)
					}
				}
			}
		})
	}
}

func TestGetTemplateContent(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		expectError  bool
		expectEmpty  bool
	}{
		{
			name:         "default template",
			templateName: "default",
			expectError:  false,
			expectEmpty:  false,
		},
		{
			name:         "go template",
			templateName: "go",
			expectError:  false,
			expectEmpty:  false,
		},
		{
			name:         "node template",
			templateName: "node",
			expectError:  false,
			expectEmpty:  false,
		},
		{
			name:         "non-existent template",
			templateName: "nonexistent",
			expectError:  true,
			expectEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := GetTemplateContent(tt.templateName)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.expectEmpty {
				if content != "" {
					t.Errorf("expected empty content, got: %s", content)
				}
			} else {
				if content == "" {
					t.Errorf("expected non-empty content")
				}
				// Verify it's valid JSON-like content
				if !strings.Contains(content, `"version"`) || !strings.Contains(content, `"tasks"`) {
					t.Errorf("expected tasks.json format, got: %s", content)
				}
			}
		})
	}
}

func TestGetAvailableTemplates(t *testing.T) {
	templates := GetAvailableTemplates()
	
	expectedTemplates := []string{"default", "go", "node"}
	
	if len(templates) != len(expectedTemplates) {
		t.Errorf("expected %d templates, got %d", len(expectedTemplates), len(templates))
	}

	for _, expected := range expectedTemplates {
		found := false
		for _, template := range templates {
			if template == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected template '%s' not found in available templates", expected)
		}
	}
}

func TestWriteTemplateToFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tasks-json-cli-write-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testContent := `{"test": "content"}`
	testFile := filepath.Join(tempDir, "test.json")

	err = writeTemplateToFile(testFile, testContent)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify file was written correctly
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read written file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("expected content '%s', got '%s'", testContent, string(content))
	}
}