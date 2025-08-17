package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/garaemon/tasks-json-cli/internal/config"
	"github.com/garaemon/tasks-json-cli/internal/discovery"
	"github.com/garaemon/tasks-json-cli/internal/executor"
	"github.com/spf13/cobra"
)

var watchPaths []string
var watchExtensions []string
var watchExclude []string
var watchDelay time.Duration

var watchCommand = &cobra.Command{
	Use:   "watch <task-name>",
	Short: "Watch files and auto-execute task",
	Long:  `Watch for file changes and automatically execute the specified task when changes are detected.`,
	Args:  cobra.ExactArgs(1),
	RunE:  executeWatchCommand,
	SilenceUsage: true,
}

func executeWatchCommand(cmd *cobra.Command, args []string) error {
	taskName := args[0]

	// Determine workspace folder
	var workspaceDir string
	if workspaceFolder != "" {
		workspaceDir = workspaceFolder
	} else {
		// Default to git root
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		gitRoot, err := discovery.FindGitRoot(currentDir)
		if err != nil {
			// Fall back to current directory if git root not found
			workspaceDir = currentDir
		} else {
			workspaceDir = gitRoot
		}
	}

	if verbose {
		fmt.Printf("Workspace folder: %s\n", workspaceDir)
	}

	tasksFilePath, err := discovery.FindTasksFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to find tasks file: %w", err)
	}

	if verbose {
		fmt.Printf("Using tasks file: %s\n", tasksFilePath)
	}

	tasks, err := config.LoadTasks(tasksFilePath)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	var targetTask *config.Task
	for i := range tasks {
		if tasks[i].Label == taskName {
			targetTask = &tasks[i]
			break
		}
	}

	if targetTask == nil {
		return fmt.Errorf("task '%s' not found", taskName)
	}

	// Set up file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	// Set default watch paths if none specified
	if len(watchPaths) == 0 {
		watchPaths = []string{workspaceDir}
	}

	// Add watch paths
	for _, path := range watchPaths {
		var fullPath string
		if filepath.IsAbs(path) {
			fullPath = path
		} else {
			fullPath = filepath.Join(workspaceDir, path)
		}

		err = addWatchPath(watcher, fullPath)
		if err != nil {
			return fmt.Errorf("failed to watch path %s: %w", fullPath, err)
		}
		if verbose {
			fmt.Printf("Watching: %s\n", fullPath)
		}
	}

	if !quiet {
		fmt.Printf("Watching for changes... (task: %s)\n", taskName)
		fmt.Println("Press Ctrl+C to stop")
	}

	// Channel for debouncing file events
	debounceTimer := time.NewTimer(0)
	debounceTimer.Stop()

	executeTask := func() {
		if !quiet {
			fmt.Printf("Executing task: %s\n", targetTask.Label)
		}
		err := executor.RunTask(targetTask, workspaceDir, file)
		if err != nil {
			log.Printf("Task execution failed: %v", err)
		}
	}

	// Watch for events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Check if we should handle this event
			if !shouldHandleEvent(event) {
				continue
			}

			if verbose {
				fmt.Printf("File changed: %s\n", event.Name)
			}

			// Debounce events - reset timer
			debounceTimer.Stop()
			debounceTimer = time.AfterFunc(watchDelay, executeTask)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watch error: %v", err)
		}
	}
}

func addWatchPath(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that should be excluded
		if info.IsDir() {
			for _, exclude := range watchExclude {
				if strings.Contains(walkPath, exclude) {
					return filepath.SkipDir
				}
			}
			return watcher.Add(walkPath)
		}

		return nil
	})
}

func shouldHandleEvent(event fsnotify.Event) bool {
	// Only handle write and create events
	if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
		return false
	}

	// Check file extensions if specified
	if len(watchExtensions) > 0 {
		ext := filepath.Ext(event.Name)
		found := false
		for _, allowedExt := range watchExtensions {
			if strings.EqualFold(ext, allowedExt) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check if file should be excluded
	for _, exclude := range watchExclude {
		if strings.Contains(event.Name, exclude) {
			return false
		}
	}

	return true
}

func init() {
	watchCommand.Flags().StringVar(&workspaceFolder, "workspace-folder", "", "workspace folder path (defaults to git root)")
	watchCommand.Flags().StringVar(&file, "file", "", "file path to replace ${file} variable")
	watchCommand.Flags().StringSliceVar(&watchPaths, "path", []string{}, "paths to watch (defaults to workspace folder)")
	watchCommand.Flags().StringSliceVar(&watchExtensions, "ext", []string{}, "file extensions to watch (e.g., .go,.js)")
	watchCommand.Flags().StringSliceVar(&watchExclude, "exclude", []string{"node_modules", ".git", ".vscode"}, "paths to exclude from watching")
	watchCommand.Flags().DurationVar(&watchDelay, "delay", 500*time.Millisecond, "delay before executing task after file change")
	rootCmd.AddCommand(watchCommand)
}