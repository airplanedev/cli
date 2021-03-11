package api

import "encoding/json"

// CreateTaskRequest creates a new task.
type CreateTaskRequest struct {
	Name           string            `json:"name" yaml:"name"`
	Description    string            `json:"description" yaml:"description"`
	Image          string            `json:"image" yaml:"image"`
	Command        []string          `json:"command" yaml:"command"`
	Arguments      []string          `json:"arguments" yaml:"arguments"`
	Parameters     Parameters        `json:"parameters" yaml:"parameters"`
	Constraints    interface{}       `json:"constraints" yaml:"constraints"`
	Env            map[string]string `json:"env" yaml:"env"`
	ResourceLimits map[string]string `json:"resourceLimits" yaml:"resourceLimits"`
	Builder        string            `json:"builder" yaml:"builder"`
	BuilderConfig  map[string]string `json:"builderConfig" yaml:"builderConfig"`
	Repo           string            `json:"repo" yaml:"repo"`
	// TODO(amir): friendly type here (120s, 5m ...)
	Timeout int `json:"timeout" yaml:"timeout"`
}

// Parameters represents a slice of task parameters.
type Parameters []Parameter

// MarshalJSON implementation.
//
// Marshals the slice of parameters as an object
// of `{ "parameters": [] }`.
//
// TODO(amir): remove once the API accepts a flat array of parameters.
func (p Parameters) MarshalJSON() ([]byte, error) {
	type object struct {
		Parameters []Parameter `json:"parameters"`
	}
	return json.Marshal(object{p})
}

// Parameter represents a task parameter.
type Parameter struct {
	Name        string      `json:"name" yaml:"name"`
	Slug        string      `json:"slug" yaml:"slug"`
	Type        string      `json:"type" yaml:"type"`
	Desc        string      `json:"desc" yaml:"desc"`
	Component   Component   `json:"component" yaml:"component"`
	Default     Value       `json:"default" yaml:"default"`
	Constraints Constraints `json:"constraints" yaml:"constraints"`
}

// Constraints represent constraints.
type Constraints struct {
	Optional bool   `json:"optional" yaml:"optional"`
	Regex    string `json:"regex" yaml:"regex"`
}

// Value represents a value.
type Value interface{}

// Component enumerates components.
type Component string

// All Component types.
const (
	ComponentNone      Component = ""
	ComponentEditorSQL Component = "editor-sql"
	ComponentTextarea  Component = "textarea"
)

// RunConstraints represents run constraints.
type RunConstraints struct {
	Labels []AgentLabel `json:"labels" yaml:"labels"`
}

// AgentLabel represents an agent label.
type AgentLabel struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

// CreateTaskResponse represents a create task response.
type CreateTaskResponse struct {
	TaskID string `json:"taskID"`
	Slug   string `json:"slug"`
}

// ListTasksResponse represents a list tasks response.
type ListTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

// Task represents a task.
//
// Even though the task object contains many other fields
// we don't add them here unless we need them for presenting tasks.
type Task struct {
	ID   string `json:"taskID"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
