package definitions

import (
	"github.com/airplanedev/cli/pkg/api"
)

type Definition_0_2 struct {
	Slug             string               `yaml:"slug"`
	Name             string               `yaml:"name"`
	Description      string               `yaml:"description,omitempty"`
	Arguments        []string             `yaml:"arguments,omitempty"`
	Parameters       api.Parameters       `yaml:"parameters,omitempty"`
	Constraints      api.RunConstraints   `yaml:"constraints,omitempty"`
	Env              api.TaskEnv          `yaml:"env,omitempty"`
	ResourceRequests api.ResourceRequests `yaml:"resourceRequests,omitempty"`
	Resources        api.Resources        `yaml:"resources,omitempty"`
	Repo             string               `yaml:"repo,omitempty"`
	Timeout          int                  `yaml:"timeout,omitempty"`

	Manual     *ManualDefinition `yaml:"manual,omitempty"`
	Deno       *DenoDefinition   `yaml:"deno,omitempty"`
	Dockerfile *DockerDefinition `yaml:"dockerfile,omitempty"`
	Go         *GoDefinition     `yaml:"go,omitempty"`
	Node       *NodeDefinition   `yaml:"node,omitempty"`
	Python     *PythonDefinition `yaml:"python,omitempty"`

	SQL *SQLDefinition `yaml:"sql,omitempty"`

	// Root is a directory path relative to the parent directory of this
	// task definition which defines what directory should be included
	// in the task's Docker image.
	//
	// If not set, defaults to "." (in other words, the parent directory of this task definition).
	//
	// This field is ignored when using the pre-built image builder (aka "manual").
	Root string `yaml:"root,omitempty"`
}

type ManualDefinition struct {
	Image   string   `yaml:"image,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

type DenoDefinition struct {
	Entrypoint string `yaml:"entrypoint"`
}

type DockerDefinition struct {
	Dockerfile string `yaml:"dockerfile"`
}

type GoDefinition struct {
	Entrypoint string `yaml:"entrypoint"`
}

type NodeDefinition struct {
	Entrypoint  string `yaml:"entrypoint"`
	Language    string `yaml:"language"`
	NodeVersion string `yaml:"nodeVersion"`
}

type PythonDefinition struct {
	Entrypoint string `yaml:"entrypoint"`
}

type SQLDefinition struct {
	Query string `yaml:"query"`
}

func (d Definition_0_2) upgrade() (Definition, error) {
	return Definition(d), nil
}
