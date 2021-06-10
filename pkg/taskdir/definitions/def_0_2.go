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

	Deno       *DenoDefinition       `yaml:"deno,omitempty"`
	Image      *ImageDefinition      `yaml:"image,omitempty"`
	Dockerfile *DockerfileDefinition `yaml:"dockerfile,omitempty"`
	Go         *GoDefinition         `yaml:"go,omitempty"`
	Node       *NodeDefinition       `yaml:"node,omitempty"`
	Python     *PythonDefinition     `yaml:"python,omitempty"`

	SQL  *SQLDefinition  `yaml:"sql,omitempty"`
	REST *RESTDefinition `yaml:"rest,omitempty"`

	// Root is a directory path relative to the parent directory of this
	// task definition which defines what directory should be included
	// in the task's Docker image.
	//
	// If not set, defaults to "." (in other words, the parent directory of this task definition).
	//
	// This field is ignored when using the "image" builder.
	Root string `yaml:"root,omitempty"`
}

type ImageDefinition struct {
	Image   string   `yaml:"image,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

type DenoDefinition struct {
	Entrypoint string `yaml:"entrypoint" mapstructure:"entrypoint"`
}

type DockerfileDefinition struct {
	Dockerfile string `yaml:"dockerfile" mapstructure:"dockerfile"`
}

type GoDefinition struct {
	Entrypoint string `yaml:"entrypoint" mapstructure:"entrypoint"`
}

type NodeDefinition struct {
	Workdir     string `yaml:"workdir" mapstructure:"workdir"`
	Entrypoint  string `yaml:"entrypoint" mapstructure:"entrypoint"`
	Language    string `yaml:"language" mapstructure:"language"`
	NodeVersion string `yaml:"nodeVersion" mapstructure:"nodeVersion"`
}

type PythonDefinition struct {
	Entrypoint string `yaml:"entrypoint" mapstructure:"entrypoint"`
}

type SQLDefinition struct {
	Query string `yaml:"query" mapstructure:"query"`
}

type RESTDefinition struct {
	Headers   map[string]string `yaml:"headers,omitempty" mapstructure:"headers"`
	Method    string            `yaml:"method" mapstructure:"method"`
	Path      string            `yaml:"path" mapstructure:"path"`
	URLParams map[string]string `yaml:"urlParams,omitempty" mapstructure:"urlParams"`
	Body      string            `yaml:"body,omitempty" mapstructure:"body,omitempty"`
	JSONBody  interface{}       `yaml:"jsonBody,omitempty" mapstructure:"jsonBody,omitempty"`
}

func (d Definition_0_2) upgrade() (Definition, error) {
	return Definition(d), nil
}
