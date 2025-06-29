package discovery

import (
	"fmt"
	"os"
	"path/filepath"
)

func findTasksFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return searchTasksFile(cwd)
}

func searchTasksFile(startDir string) (string, error) {
	dir := startDir

	for {
		tasksPath := filepath.Join(dir, ".vscode", "tasks.json")
		if _, err := os.Stat(tasksPath); err == nil {
			return tasksPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("tasks.json not found in current directory or any parent directory")
}

func FindTasksFile(configPath string) (string, error) {
	if configPath != "" {
		if _, err := os.Stat(configPath); err != nil {
			return "", fmt.Errorf("specified config file not found: %s", configPath)
		}
		return configPath, nil
	}

	return findTasksFile()
}