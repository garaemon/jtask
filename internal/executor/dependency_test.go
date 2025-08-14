package executor

import (
	"testing"

	"github.com/garaemon/tasks-json-cli/internal/config"
)

func TestNewDependencyResolver(t *testing.T) {
	tasks := []config.Task{
		{Label: "task1", Type: "shell", Command: "echo task1"},
		{Label: "task2", Type: "shell", Command: "echo task2"},
	}

	resolver := NewDependencyResolver(tasks)

	if len(resolver.tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(resolver.tasks))
	}

	if resolver.tasks["task1"].Label != "task1" {
		t.Error("Task1 not found in resolver")
	}

	if resolver.tasks["task2"].Label != "task2" {
		t.Error("Task2 not found in resolver")
	}
}

func TestResolveExecutionOrderNoDependencies(t *testing.T) {
	tasks := []config.Task{
		{Label: "task1", Type: "shell", Command: "echo task1"},
	}

	resolver := NewDependencyResolver(tasks)
	order, err := resolver.ResolveExecutionOrder("task1")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(order) != 1 {
		t.Errorf("Expected 1 task, got %d", len(order))
	}

	if order[0].Label != "task1" {
		t.Errorf("Expected task1, got %s", order[0].Label)
	}
}

func TestResolveExecutionOrderWithDependencies(t *testing.T) {
	tasks := []config.Task{
		{Label: "build", Type: "shell", Command: "echo building"},
		{Label: "test", Type: "shell", Command: "echo testing", DependsOn: "build"},
	}

	resolver := NewDependencyResolver(tasks)
	order, err := resolver.ResolveExecutionOrder("test")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(order) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(order))
	}

	if order[0].Label != "build" {
		t.Errorf("Expected build to be first, got %s", order[0].Label)
	}

	if order[1].Label != "test" {
		t.Errorf("Expected test to be second, got %s", order[1].Label)
	}
}

func TestResolveExecutionOrderWithMultipleDependencies(t *testing.T) {
	tasks := []config.Task{
		{Label: "compile", Type: "shell", Command: "echo compiling"},
		{Label: "lint", Type: "shell", Command: "echo linting"},
		{Label: "test", Type: "shell", Command: "echo testing", DependsOn: []interface{}{"compile", "lint"}},
	}

	resolver := NewDependencyResolver(tasks)
	order, err := resolver.ResolveExecutionOrder("test")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(order) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(order))
	}

	if order[2].Label != "test" {
		t.Errorf("Expected test to be last, got %s", order[2].Label)
	}

	compileFound := false
	lintFound := false
	for i := 0; i < 2; i++ {
		if order[i].Label == "compile" {
			compileFound = true
		}
		if order[i].Label == "lint" {
			lintFound = true
		}
	}

	if !compileFound {
		t.Error("compile task not found in dependencies")
	}
	if !lintFound {
		t.Error("lint task not found in dependencies")
	}
}

func TestResolveExecutionOrderCircularDependency(t *testing.T) {
	tasks := []config.Task{
		{Label: "task1", Type: "shell", Command: "echo task1", DependsOn: "task2"},
		{Label: "task2", Type: "shell", Command: "echo task2", DependsOn: "task1"},
	}

	resolver := NewDependencyResolver(tasks)
	_, err := resolver.ResolveExecutionOrder("task1")

	if err == nil {
		t.Error("Expected circular dependency error")
	}

	if err.Error() != "circular dependency detected: task 'task1' depends on itself" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestResolveExecutionOrderMissingTask(t *testing.T) {
	tasks := []config.Task{
		{Label: "task1", Type: "shell", Command: "echo task1", DependsOn: "nonexistent"},
	}

	resolver := NewDependencyResolver(tasks)
	_, err := resolver.ResolveExecutionOrder("task1")

	if err == nil {
		t.Error("Expected missing task error")
	}

	if err.Error() != "task 'nonexistent' not found" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestGetMissingDependencies(t *testing.T) {
	tasks := []config.Task{
		{Label: "task1", Type: "shell", Command: "echo task1", DependsOn: "missing1"},
		{Label: "task2", Type: "shell", Command: "echo task2", DependsOn: []interface{}{"missing2", "task1"}},
	}

	resolver := NewDependencyResolver(tasks)
	missing := resolver.GetMissingDependencies()

	if len(missing) != 2 {
		t.Errorf("Expected 2 missing dependencies, got %d", len(missing))
	}

	expectedMissing := map[string]bool{"missing1": true, "missing2": true}
	for _, dep := range missing {
		if !expectedMissing[dep] {
			t.Errorf("Unexpected missing dependency: %s", dep)
		}
	}
}

func TestValidateDependencies(t *testing.T) {
	tasks := []config.Task{
		{Label: "build", Type: "shell", Command: "echo building"},
		{Label: "test", Type: "shell", Command: "echo testing", DependsOn: "build"},
	}

	resolver := NewDependencyResolver(tasks)
	err := resolver.ValidateDependencies()

	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestValidateDependenciesWithCircular(t *testing.T) {
	tasks := []config.Task{
		{Label: "task1", Type: "shell", Command: "echo task1", DependsOn: "task2"},
		{Label: "task2", Type: "shell", Command: "echo task2", DependsOn: "task1"},
	}

	resolver := NewDependencyResolver(tasks)
	err := resolver.ValidateDependencies()

	if err == nil {
		t.Error("Expected validation error for circular dependency")
	}
}