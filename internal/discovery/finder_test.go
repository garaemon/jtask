package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindTasksFile_WithConfigPath(t *testing.T) {
	testFile := filepath.Join("..", "..", "testdata", "simple_tasks.json")
	
	result, err := FindTasksFile(testFile)
	if err != nil {
		t.Fatalf("FindTasksFile failed: %v", err)
	}
	
	if result != testFile {
		t.Errorf("Expected %s, got %s", testFile, result)
	}
}

func TestFindTasksFile_ConfigPathNotExists(t *testing.T) {
	_, err := FindTasksFile("nonexistent.json")
	if err == nil {
		t.Fatal("Expected error for nonexistent config file")
	}
}

func TestSearchTasksFile(t *testing.T) {
	tempDir := t.TempDir()
	
	vscodePath := filepath.Join(tempDir, ".vscode")
	err := os.MkdirAll(vscodePath, 0755)
	if err != nil {
		t.Fatalf("Failed to create .vscode directory: %v", err)
	}
	
	tasksFile := filepath.Join(vscodePath, "tasks.json")
	err = os.WriteFile(tasksFile, []byte(`{"version": "2.0.0", "tasks": []}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create tasks.json: %v", err)
	}
	
	result, err := searchTasksFile(tempDir)
	if err != nil {
		t.Fatalf("searchTasksFile failed: %v", err)
	}
	
	if result != tasksFile {
		t.Errorf("Expected %s, got %s", tasksFile, result)
	}
}

func TestSearchTasksFile_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	
	_, err := searchTasksFile(tempDir)
	if err == nil {
		t.Fatal("Expected error when tasks.json not found")
	}
}

func TestSearchTasksFile_ParentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	
	vscodePath := filepath.Join(tempDir, ".vscode")
	err := os.MkdirAll(vscodePath, 0755)
	if err != nil {
		t.Fatalf("Failed to create .vscode directory: %v", err)
	}
	
	tasksFile := filepath.Join(vscodePath, "tasks.json")
	err = os.WriteFile(tasksFile, []byte(`{"version": "2.0.0", "tasks": []}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create tasks.json: %v", err)
	}
	
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	result, err := searchTasksFile(subDir)
	if err != nil {
		t.Fatalf("searchTasksFile failed: %v", err)
	}
	
	if result != tasksFile {
		t.Errorf("Expected %s, got %s", tasksFile, result)
	}
}