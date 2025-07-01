package discovery

import (
	"os"
	"path/filepath"
)

// FindGitRoot recursively searches for git root directory starting from the given path
func FindGitRoot(startPath string) (string, error) {
	currentPath := startPath
	
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return currentPath, nil
		}
		
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory without finding .git
			return "", os.ErrNotExist
		}
		currentPath = parentPath
	}
}