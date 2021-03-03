package api

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

// Parameter represents a task parameter.
type Parameter struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	Format     string `json:"format" yaml:"format"`
	Label      string `json:"label" yaml:"label"`
	HelpText   string `json:"helpText" yaml:"helpText"`
	Default    string `json:"default" yaml:"default"`
	TrueValue  string `json:"trueValue" yaml:"trueValue"`
	FalseValue string `json:"falseValue" yaml:"falseValue"`
}

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
}
