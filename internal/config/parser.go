package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tidwall/jsonc"
)

func parseTasksFile(filePath string) (*TasksFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	// Parse JSONC (JSON with comments and trailing commas)
	jsonData := jsonc.ToJSON(data)
	
	var tasksFile TasksFile
	if err := json.Unmarshal(jsonData, &tasksFile); err != nil {
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