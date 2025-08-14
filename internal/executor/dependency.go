package executor

import (
	"fmt"
	"sort"

	"github.com/garaemon/tasks-json-cli/internal/config"
)

type DependencyResolver struct {
	tasks map[string]*config.Task
}

func NewDependencyResolver(tasks []config.Task) *DependencyResolver {
	taskMap := make(map[string]*config.Task)
	for i := range tasks {
		taskMap[tasks[i].Label] = &tasks[i]
	}
	return &DependencyResolver{tasks: taskMap}
}

func (r *DependencyResolver) ResolveExecutionOrder(taskLabel string) ([]*config.Task, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []*config.Task
	
	err := r.visitTask(taskLabel, visited, visiting, &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

func (r *DependencyResolver) visitTask(taskLabel string, visited, visiting map[string]bool, result *[]*config.Task) error {
	if visiting[taskLabel] {
		return fmt.Errorf("circular dependency detected: task '%s' depends on itself", taskLabel)
	}
	
	if visited[taskLabel] {
		return nil
	}
	
	task, exists := r.tasks[taskLabel]
	if !exists {
		return fmt.Errorf("task '%s' not found", taskLabel)
	}
	
	visiting[taskLabel] = true
	
	dependencies := task.GetDependencies()
	dependsOrder := task.GetDependsOrder()
	
	if dependsOrder == "sequence" {
		for _, dep := range dependencies {
			err := r.visitTask(dep, visited, visiting, result)
			if err != nil {
				return err
			}
		}
	} else {
		for _, dep := range dependencies {
			err := r.visitTask(dep, visited, visiting, result)
			if err != nil {
				return err
			}
		}
	}
	
	visiting[taskLabel] = false
	visited[taskLabel] = true
	
	*result = append(*result, task)
	return nil
}

func (r *DependencyResolver) GetParallelGroups(taskLabel string) ([][]*config.Task, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var groups [][]*config.Task
	
	err := r.buildParallelGroups(taskLabel, visited, visiting, &groups)
	if err != nil {
		return nil, err
	}
	
	return groups, nil
}

func (r *DependencyResolver) buildParallelGroups(taskLabel string, visited, visiting map[string]bool, groups *[][]*config.Task) error {
	if visiting[taskLabel] {
		return fmt.Errorf("circular dependency detected: task '%s' depends on itself", taskLabel)
	}
	
	if visited[taskLabel] {
		return nil
	}
	
	task, exists := r.tasks[taskLabel]
	if !exists {
		return fmt.Errorf("task '%s' not found", taskLabel)
	}
	
	visiting[taskLabel] = true
	
	dependencies := task.GetDependencies()
	dependsOrder := task.GetDependsOrder()
	
	if len(dependencies) > 0 {
		if dependsOrder == "sequence" {
			for _, dep := range dependencies {
				err := r.buildParallelGroups(dep, visited, visiting, groups)
				if err != nil {
					return err
				}
			}
		} else {
			var parallelTasks []*config.Task
			for _, dep := range dependencies {
				err := r.buildParallelGroups(dep, visited, visiting, groups)
				if err != nil {
					return err
				}
				if depTask, exists := r.tasks[dep]; exists {
					parallelTasks = append(parallelTasks, depTask)
				}
			}
			if len(parallelTasks) > 0 {
				*groups = append(*groups, parallelTasks)
			}
		}
	}
	
	visiting[taskLabel] = false
	visited[taskLabel] = true
	
	*groups = append(*groups, []*config.Task{task})
	return nil
}

func (r *DependencyResolver) ValidateDependencies() error {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	
	for taskLabel := range r.tasks {
		if !visited[taskLabel] {
			var result []*config.Task
			err := r.visitTask(taskLabel, visited, visiting, &result)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (r *DependencyResolver) GetMissingDependencies() []string {
	var missing []string
	
	for _, task := range r.tasks {
		dependencies := task.GetDependencies()
		for _, dep := range dependencies {
			if _, exists := r.tasks[dep]; !exists {
				missing = append(missing, dep)
			}
		}
	}
	
	sort.Strings(missing)
	return missing
}