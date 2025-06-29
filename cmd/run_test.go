package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecuteRunCommand_TaskNotFound(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	
	cmd := &cobra.Command{}
	args := []string{"nonexistent"}
	
	err := executeRunCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for non-existent task")
	}
	
	if err.Error() != "task 'nonexistent' not found" {
		t.Errorf("expected error message about task not found, got %s", err.Error())
	}
}

func TestExecuteRunCommand_DryRun(t *testing.T) {
	configPath = "../testdata/simple_tasks.json"
	dryRun = true
	defer func() { dryRun = false }()
	
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	_ = w.Close()
	out, _ := io.ReadAll(r)
	
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	output := string(out)
	if !bytes.Contains(out, []byte("Would execute task: build")) {
		t.Errorf("expected dry run output, got %s", output)
	}
}

func TestExecuteRunCommand_InvalidTasksFile(t *testing.T) {
	configPath = "../testdata/invalid_tasks.json"
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for invalid tasks file")
	}
}

func TestExecuteRunCommand_TasksFileNotFound(t *testing.T) {
	configPath = "nonexistent.json"
	
	cmd := &cobra.Command{}
	args := []string{"build"}
	
	err := executeRunCommand(cmd, args)
	
	if err == nil {
		t.Error("expected error for non-existent tasks file")
	}
}