package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func parseTasksFile(filePath string) (*TasksFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	var tasksFile TasksFile
	if err := json.Unmarshal(data, &tasksFile); err != nil {
		return nil, fmt.Errorf("failed to parse tasks file: %w", err)
	}

	return &tasksFile, nil
}

func LoadTasks(filePath string) ([]Task, error) {
	tasksFile, err := parseTasksFile(filePath)
	if err != nil {
		return nil, err
	}

	return tasksFile.Tasks, nil
}