package config

type TasksFile struct {
	Version string `json:"version"`
	Tasks   []Task `json:"tasks"`
}

type Task struct {
	Label           string            `json:"label"`
	Type            string            `json:"type"`
	Command         string            `json:"command"`
	Args            []string          `json:"args,omitempty"`
	Group           interface{}       `json:"group,omitempty"`
	ProblemMatcher  interface{}       `json:"problemMatcher,omitempty"`
	Options         *TaskOptions      `json:"options,omitempty"`
	DependsOn       interface{}       `json:"dependsOn,omitempty"`
	DependsOrder    string            `json:"dependsOrder,omitempty"`
	Presentation    *TaskPresentation `json:"presentation,omitempty"`
	RunOptions      *TaskRunOptions   `json:"runOptions,omitempty"`
}

type TaskOptions struct {
	Cwd   string            `json:"cwd,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
	Shell *ShellOptions     `json:"shell,omitempty"`
}

type ShellOptions struct {
	Executable string   `json:"executable,omitempty"`
	Args       []string `json:"args,omitempty"`
}

type TaskPresentation struct {
	Echo       *bool  `json:"echo,omitempty"`
	Reveal     string `json:"reveal,omitempty"`
	Focus      *bool  `json:"focus,omitempty"`
	Panel      string `json:"panel,omitempty"`
	ShowReuseMessage *bool `json:"showReuseMessage,omitempty"`
	Clear      *bool  `json:"clear,omitempty"`
	Group      string `json:"group,omitempty"`
}

type TaskRunOptions struct {
	RunOn string `json:"runOn,omitempty"`
}

func (t *Task) GetGroupKind() string {
	if t.Group == nil {
		return ""
	}
	
	switch group := t.Group.(type) {
	case string:
		return group
	case map[string]interface{}:
		if kind, ok := group["kind"].(string); ok {
			return kind
		}
	}
	
	return ""
}

func (t *Task) IsDefaultInGroup() bool {
	if t.Group == nil {
		return false
	}
	
	if group, ok := t.Group.(map[string]interface{}); ok {
		if isDefault, ok := group["isDefault"].(bool); ok {
			return isDefault
		}
	}
	
	return false
}