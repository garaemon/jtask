package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	// Create a temporary directory structure
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

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test finding git root from subdirectory
	gitRoot, err := FindGitRoot(subDir)
	if err != nil {
		t.Fatalf("Expected to find git root, got error: %v", err)
	}

	if gitRoot != tempDir {
		t.Errorf("Expected git root to be %s, got %s", tempDir, gitRoot)
	}

	// Test finding git root from git directory itself
	gitRoot, err = FindGitRoot(tempDir)
	if err != nil {
		t.Fatalf("Expected to find git root, got error: %v", err)
	}

	if gitRoot != tempDir {
		t.Errorf("Expected git root to be %s, got %s", tempDir, gitRoot)
	}
}

func TestFindGitRootNotFound(t *testing.T) {
	// Create a temporary directory without .git
	tempDir, err := os.MkdirTemp("", "no-git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	_, err = FindGitRoot(tempDir)
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}