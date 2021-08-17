package api

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

// CreateTaskRequest creates a new task.
type CreateTaskRequest struct {
	Slug             string            `json:"slug"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Image            *string           `json:"image"`
	Command          []string          `json:"command"`
	Arguments        []string          `json:"arguments"`
	Parameters       Parameters        `json:"parameters"`
	Constraints      RunConstraints    `json:"constraints"`
	Env              TaskEnv           `json:"env"`
	ResourceRequests map[string]string `json:"resourceRequests"`
	Resources        map[string]string `json:"resources"`
	Kind             TaskKind          `json:"kind"`
	KindOptions      KindOptions       `json:"kindOptions"`
	Repo             string            `json:"repo"`
	// TODO(amir): friendly type here (120s, 5m ...)
	Timeout int `json:"timeout"`
}

// UpdateTaskRequest updates a task.
type UpdateTaskRequest struct {
	Slug                       string            `json:"slug" yaml:"-"`
	Name                       string            `json:"name" yaml:"name"`
	Description                string            `json:"description" yaml:"description"`
	Image                      *string           `json:"image" yaml:"image"`
	Command                    []string          `json:"command" yaml:"command"`
	Arguments                  []string          `json:"arguments" yaml:"arguments"`
	Parameters                 Parameters        `json:"parameters" yaml:"parameters"`
	Constraints                RunConstraints    `json:"constraints" yaml:"constraints"`
	Env                        TaskEnv           `json:"env" yaml:"env"`
	ResourceRequests           map[string]string `json:"resourceRequests" yaml:"resourceRequests"`
	Resources                  map[string]string `json:"resources" yaml:"resources"`
	Kind                       TaskKind          `json:"kind" yaml:"builder"`
	KindOptions                KindOptions       `json:"kindOptions" yaml:"builderConfig"`
	Repo                       string            `json:"repo" yaml:"repo"`
	RequireExplicitPermissions bool              `json:"requireExplicitPermissions" yaml:"-"`
	Permissions                Permissions       `json:"permissions" yaml:"-"`
	// TODO(amir): friendly type here (120s, 5m ...)
	Timeout int `json:"timeout" yaml:"timeout"`
}

type Permissions []Permission

type Permission struct {
	Action     Action  `json:"action"`
	SubUserID  *string `json:"subUserID"`
	SubGroupID *string `json:"subGroupID"`
}

type Action string

type TaskKind string

const (
	TaskKindDeno       TaskKind = "deno"
	TaskKindDockerfile TaskKind = "dockerfile"
	TaskKindGo         TaskKind = "go"
	TaskKindImage      TaskKind = "image"
	TaskKindNode       TaskKind = "node"
	TaskKindPython     TaskKind = "python"
	TaskKindShell      TaskKind = "shell"

	TaskKindSQL  TaskKind = "sql"
	TaskKindREST TaskKind = "rest"
)

type UpdateTaskResponse struct {
	TaskRevisionID string `json:"taskRevisionID"`
}

// GetLogsResponse represents a get logs response.
type GetLogsResponse struct {
	RunID string    `json:"runID"`
	Logs  []LogItem `json:"logs"`
}

// GetBuildLogsResponse represents a get build logs response.
type GetBuildLogsResponse struct {
	BuildID string    `json:"buildID"`
	Logs    []LogItem `json:"logs"`
}

// Outputs represents outputs.
type Outputs map[string][]interface{}

// Represents a line of the output
type OutputRow struct {
	OutputName string      `json:"name" yaml:"name"`
	Value      interface{} `json:"value" yaml:"value"`
}

// GetOutputsResponse represents a get outputs response.
type GetOutputsResponse struct {
	Outputs Outputs `json:"outputs"`
}

// LogItem represents a log item.
type LogItem struct {
	Timestamp time.Time `json:"timestamp"`
	InsertID  string    `json:"insertID"`
	Text      string    `json:"text"`
	Level     LogLevel  `json:"level"`
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
)

// RegistryTokenResponse represents a registry token response.
type RegistryTokenResponse struct {
	Token      string `json:"token"`
	Expiration string `json:"expiration"`
	Repo       string `json:"repo"`
}

// Parameters represents a slice of task parameters.
//
// TODO(amir): remove custom marshal/unmarshal once the API is updated.
type Parameters []Parameter

// UnmarshalJSON implementation.
func (p *Parameters) UnmarshalJSON(buf []byte) error {
	var tmp struct {
		Parameters []Parameter `json:"parameters"`
	}

	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}

	*p = tmp.Parameters
	return nil
}

// MarshalJSON implementation.
func (p Parameters) MarshalJSON() ([]byte, error) {
	type object struct {
		Parameters []Parameter `json:"parameters"`
	}
	return json.Marshal(object{p})
}

// Type enumerates parameter types.
type Type string

// All Parameter types.
const (
	TypeString   Type = "string"
	TypeBoolean  Type = "boolean"
	TypeUpload   Type = "upload"
	TypeInteger  Type = "integer"
	TypeFloat    Type = "float"
	TypeDate     Type = "date"
	TypeDatetime Type = "datetime"
)

// Parameter represents a task parameter.
type Parameter struct {
	Name        string      `json:"name" yaml:"name"`
	Slug        string      `json:"slug" yaml:"slug"`
	Type        Type        `json:"type" yaml:"type"`
	Desc        string      `json:"desc" yaml:"desc,omitempty"`
	Component   Component   `json:"component" yaml:"component,omitempty"`
	Default     Value       `json:"default" yaml:"default,omitempty"`
	Constraints Constraints `json:"constraints" yaml:"constraints,omitempty"`
}

// Constraints represent constraints.
type Constraints struct {
	Optional bool   `json:"optional" yaml:"optional,omitempty"`
	Regex    string `json:"regex" yaml:"regex,omitempty"`
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

// AuthInfoResponse represents info about authenticated user.
type AuthInfoResponse struct {
	User *UserInfo `json:"user"`
	Team *TeamInfo `json:"team"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type TeamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateTaskResponse represents a create task response.
type CreateTaskResponse struct {
	TaskID         string `json:"taskID"`
	Slug           string `json:"slug"`
	TaskRevisionID string `json:"taskRevisionID"`
}

// ListTasksResponse represents a list tasks response.
type ListTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type TaskEnv map[string]EnvVarValue

type EnvVarValue struct {
	Value  *string `json:"value" yaml:"value,omitempty"`
	Config *string `json:"config" yaml:"config,omitempty"`
}

var _ yaml.Unmarshaler = &EnvVarValue{}

// UnmarshalJSON allows you set an env var's `value` using either
// of these notations:
//
//   AIRPLANE_DSN: "foobar"
//
//   AIRPLANE_DSN:
//     value: "foobar"
//
func (ev *EnvVarValue) UnmarshalYAML(node *yaml.Node) error {
	// First, try to unmarshal as a string.
	// This would be the first case above.
	var value string
	if err := node.Decode(&value); err == nil {
		// Success!
		ev.Value = &value
		return nil
	}

	// Otherwise, perform a normal unmarshal operation.
	// This would be the second case above.
	//
	// Note we need a new type, otherwise we recursively call this
	// method and end up stack overflowing.
	type envVarValue EnvVarValue
	var v envVarValue
	if err := node.Decode(&v); err != nil {
		return err
	}
	*ev = EnvVarValue(v)

	return nil
}

// Task represents a task.
type Task struct {
	URL                        string           `json:"-" yaml:"-"`
	ID                         string           `json:"taskID" yaml:"id"`
	Name                       string           `json:"name" yaml:"name"`
	Slug                       string           `json:"slug" yaml:"slug"`
	Description                string           `json:"description" yaml:"description"`
	Image                      *string          `json:"image" yaml:"image"`
	Command                    []string         `json:"command" yaml:"command"`
	Arguments                  []string         `json:"arguments" yaml:"arguments"`
	Parameters                 Parameters       `json:"parameters" yaml:"parameters"`
	Constraints                RunConstraints   `json:"constraints" yaml:"constraints"`
	Env                        TaskEnv          `json:"env" yaml:"env"`
	ResourceRequests           ResourceRequests `json:"resourceRequests" yaml:"resourceRequests"`
	Resources                  Resources        `json:"resources" yaml:"resources"`
	Kind                       TaskKind         `json:"kind" yaml:"kind"`
	KindOptions                KindOptions      `json:"kindOptions" yaml:"kindOptions"`
	Repo                       string           `json:"repo" yaml:"repo"`
	RequireExplicitPermissions bool             `json:"requireExplicitPermissions" yaml:"-"`
	Permissions                Permissions      `json:"permissions" yaml:"-"`
	Timeout                    int              `json:"timeout" yaml:"timeout"`
}

type KindOptions map[string]interface{}

type ResourceRequests map[string]string

type Resources map[string]string

type ResourceKind string

const (
	KindUnknown  ResourceKind = ""
	KindPostgres ResourceKind = "postgres"
	KindMySQL    ResourceKind = "mysql"
	KindREST     ResourceKind = "rest"
)

type Resource struct {
	ID         string                 `json:"id" db:"id"`
	TeamID     string                 `json:"teamID" db:"team_id"`
	Name       string                 `json:"name" db:"name"`
	Kind       ResourceKind           `json:"kind" db:"kind"`
	KindConfig map[string]interface{} `json:"kindConfig" db:"kind_config"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	CreatedBy string    `json:"createdBy" db:"created_by"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	UpdatedBy string    `json:"updatedBy" db:"updated_by"`

	IsPrivate bool `json:"isPrivate" db:"is_private"`
}

// Values represent parameters values.
//
// An alias is used because we want the type
// to be `map[string]interface{}` and not a custom one.
//
// They're keyed by the parameter "slug".
type Values = map[string]interface{}

// RunTaskRequest represents a run task request.
type RunTaskRequest struct {
	TaskID      string            `json:"taskID"`
	ParamValues Values            `json:"params"`
	Env         map[string]string `json:"env"`
	Constraints Constraints       `json:"constraints"`
}

// RunTaskResponse represents a run task response.
type RunTaskResponse struct {
	RunID string `json:"runID"`
}

// GetRunResponse represents a get task response.
type GetRunResponse struct {
	Run Run `json:"run"`
}

// RunStatus enumerates run status.
type RunStatus string

// All RunStatus types.
const (
	RunNotStarted RunStatus = "NotStarted"
	RunQueued     RunStatus = "Queued"
	RunActive     RunStatus = "Active"
	RunSucceeded  RunStatus = "Succeeded"
	RunFailed     RunStatus = "Failed"
	RunCancelled  RunStatus = "Cancelled"
)

// Run represents a run.
type Run struct {
	RunID       string     `json:"runID"`
	TaskID      string     `json:"taskID"`
	TaskName    string     `json:"taskName"`
	TeamID      string     `json:"teamID"`
	Status      RunStatus  `json:"status"`
	ParamValues Values     `json:"paramValues"`
	CreatedAt   time.Time  `json:"createdAt"`
	CreatorID   string     `json:"creatorID"`
	QueuedAt    *time.Time `json:"queuedAt"`
	ActiveAt    *time.Time `json:"activeAt"`
	SucceededAt *time.Time `json:"succeededAt"`
	FailedAt    *time.Time `json:"failedAt"`
	CancelledAt *time.Time `json:"cancelledAt"`
	CancelledBy *string    `json:"cancelledBy"`
}

// ListRunsRequest represents a list runs request.
type ListRunsRequest struct {
	TaskID string    `json:"taskID"`
	Since  time.Time `json:"since"`
	Until  time.Time `json:"until"`
	Page   int       `json:"page"`
	Limit  int       `json:"limit"`
}

// ListRunsResponse represents a list runs response.
type ListRunsResponse struct {
	Runs []Run `json:"runs"`
}

// GetConfigRequest represents a get config request
type GetConfigRequest struct {
	Name       string `json:"name"`
	Tag        string `json:"tag"`
	ShowSecret bool   `json:"showSecret"`
}

// SetConfigRequest represents a set config request.
type SetConfigRequest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

// Config represents a config var.
type Config struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

// GetConfigResponse represents a get config response.
type GetConfigResponse struct {
	Config Config `json:"config"`
}

type GetBuildResponse struct {
	Build Build `json:"build"`
}

type CreateBuildRequest struct {
	TaskID         string  `json:"taskID"`
	SourceUploadID string  `json:"sourceUploadID"`
	Env            TaskEnv `json:"env"`
}

type CreateBuildResponse struct {
	Build Build `json:"build"`
}

type Build struct {
	ID             string      `json:"id"`
	TaskRevisionID string      `json:"taskRevisionID"`
	Status         BuildStatus `json:"status"`
	CreatedAt      time.Time   `json:"createdAt"`
	CreatorID      string      `json:"creatorID"`
	QueuedAt       *time.Time  `json:"queuedAt"`
	QueuedBy       *string     `json:"queuedBy"`
	SourceUploadID string      `json:"sourceUploadID"`
}

type BuildStatus string

const (
	BuildNotStarted BuildStatus = "NotStarted"
	BuildActive     BuildStatus = "Active"
	BuildSucceeded  BuildStatus = "Succeeded"
	BuildFailed     BuildStatus = "Failed"
	BuildCancelled  BuildStatus = "Cancelled"
)

func (s BuildStatus) Stopped() bool {
	return s == BuildSucceeded || s == BuildFailed || s == BuildCancelled
}

type CreateBuildUploadRequest struct {
	SizeBytes int `json:"sizeBytes"`
}

type CreateBuildUploadResponse struct {
	Upload       Upload `json:"upload"`
	WriteOnlyURL string `json:"writeOnlyURL"`
}

type Upload struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type CreateAPIKeyResponse struct {
	APIKey APIKey `json:"apiKey"`
}

type ListAPIKeysResponse struct {
	APIKeys []APIKey `json:"apiKeys"`
}

type DeleteAPIKeyRequest struct {
	KeyID string `json:"keyID"`
}

type APIKey struct {
	ID        string    `json:"id" yaml:"id"`
	TeamID    string    `json:"teamID" yaml:"teamID"`
	Name      string    `json:"name" yaml:"name"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	Key       string    `json:"key" yaml:"key"`
}

type GetUniqueSlugResponse struct {
	Slug string `json:"slug"`
}

type ListResourcesResponse struct {
	Resources []Resource `json:"resources"`
}
