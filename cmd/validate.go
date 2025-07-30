package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/garaemon/tasks-json-cli/internal/config"
	"github.com/garaemon/tasks-json-cli/internal/discovery"
	"github.com/spf13/cobra"
)

type ValidationResult struct {
	Path     string           `json:"path"`
	Valid    bool             `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

type ValidationError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
	TaskLabel   string `json:"task_label,omitempty"`
}

var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate tasks.json syntax",
	Long:  `Validate the syntax and structure of tasks.json configuration files.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runValidateCommand,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidateCommand(cmd *cobra.Command, args []string) error {
	var targetPath string
	
	if len(args) > 0 {
		targetPath = args[0]
	} else {
		var err error
		targetPath, err = discovery.FindTasksFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to find tasks file: %w", err)
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Validating tasks file: %s\n", targetPath)
	}

	result := validateTasksFile(targetPath)
	
	if quiet {
		if !result.Valid {
			os.Exit(1)
		}
		return nil
	}

	printValidationResult(result)
	
	if !result.Valid {
		os.Exit(1)
	}
	
	return nil
}

func validateTasksFile(path string) ValidationResult {
	result := ValidationResult{
		Path:     path,
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "file_not_found",
			Message: fmt.Sprintf("tasks.json file not found: %s", path),
		})
		return result
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "read_error",
			Message: fmt.Sprintf("failed to read file: %v", err),
		})
		return result
	}

	// Validate JSON syntax
	var jsonData interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "json_syntax",
			Message: fmt.Sprintf("invalid JSON syntax: %v", err),
		})
		return result
	}

	// Validate tasks.json structure
	tasks, err := config.LoadTasks(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "structure_error",
			Message: fmt.Sprintf("invalid tasks.json structure: %v", err),
		})
		return result
	}

	// Validate individual tasks
	validateTasks(tasks, &result)

	return result
}

func validateTasks(tasks []config.Task, result *ValidationResult) {
	seenLabels := make(map[string]bool)
	
	for _, task := range tasks {
		// Check for duplicate labels
		if seenLabels[task.Label] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:      "duplicate_label",
				Message:   fmt.Sprintf("duplicate task label: %s", task.Label),
				TaskLabel: task.Label,
			})
		}
		seenLabels[task.Label] = true
		
		// Validate required fields
		if task.Label == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "missing_label",
				Message: "task is missing required 'label' field",
			})
		}
		
		if task.Type == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:      "missing_type",
				Message:   "task is missing required 'type' field",
				TaskLabel: task.Label,
			})
		}
		
		if task.Command == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:      "missing_command",
				Message:   "task is missing required 'command' field",
				TaskLabel: task.Label,
			})
		}
		
		// Validate task type
		validTypes := []string{"shell", "process"}
		isValidType := false
		for _, validType := range validTypes {
			if task.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			result.Warnings = append(result.Warnings, ValidationError{
				Type:      "unknown_type",
				Message:   fmt.Sprintf("unknown task type '%s', supported types: %v", task.Type, validTypes),
				TaskLabel: task.Label,
			})
		}
		
		// Validate working directory if specified
		if task.Options != nil && task.Options.Cwd != "" {
			// Only warn if it's an absolute path that doesn't exist
			if filepath.IsAbs(task.Options.Cwd) {
				if _, err := os.Stat(task.Options.Cwd); os.IsNotExist(err) {
					result.Warnings = append(result.Warnings, ValidationError{
						Type:      "invalid_cwd",
						Message:   fmt.Sprintf("working directory does not exist: %s", task.Options.Cwd),
						TaskLabel: task.Label,
					})
				}
			}
		}
		
		// Validate dependsOn references
		dependsOn := getDependsOnAsStringSlice(task.DependsOn)
		for _, dep := range dependsOn {
			if !seenLabels[dep] {
				// Check if the dependency exists in the full task list
				found := false
				for _, t := range tasks {
					if t.Label == dep {
						found = true
						break
					}
				}
				if !found {
					result.Warnings = append(result.Warnings, ValidationError{
						Type:      "unknown_dependency",
						Message:   fmt.Sprintf("task depends on unknown task: %s", dep),
						TaskLabel: task.Label,
					})
				}
			}
		}
	}
}

func printValidationResult(result ValidationResult) {
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Printf("✓ %s is valid\n", result.Path)
		return
	}
	
	fmt.Printf("Validation results for %s:\n", result.Path)
	fmt.Println()
	
	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			if err.TaskLabel != "" {
				fmt.Printf("  ✗ [%s] %s: %s\n", err.TaskLabel, err.Type, err.Message)
			} else {
				fmt.Printf("  ✗ %s: %s\n", err.Type, err.Message)
			}
		}
		fmt.Println()
	}
	
	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range result.Warnings {
			if warning.TaskLabel != "" {
				fmt.Printf("  ⚠ [%s] %s: %s\n", warning.TaskLabel, warning.Type, warning.Message)
			} else {
				fmt.Printf("  ⚠ %s: %s\n", warning.Type, warning.Message)
			}
		}
		fmt.Println()
	}
	
	if result.Valid {
		fmt.Printf("✓ File is valid (with %d warnings)\n", len(result.Warnings))
	} else {
		fmt.Printf("✗ File is invalid (%d errors, %d warnings)\n", len(result.Errors), len(result.Warnings))
	}
}