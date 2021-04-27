package definitions

import (
	"github.com/airplanedev/cli/pkg/api"
)

type Definition_0_2 struct {
	Slug           string             `yaml:"slug"`
	Name           string             `yaml:"name"`
	Description    string             `yaml:"description,omitempty"`
	Image          string             `yaml:"image,omitempty"`
	Command        []string           `yaml:"command,omitempty"`
	Arguments      []string           `yaml:"arguments,omitempty"`
	Parameters     api.Parameters     `yaml:"parameters,omitempty"`
	Constraints    api.RunConstraints `yaml:"constraints,omitempty"`
	Env            api.TaskEnv        `yaml:"env,omitempty"`
	ResourceLimits api.ResourceLimits `yaml:"resourceLimits,omitempty"`
	Kind           string             `yaml:"kind,omitempty"`
	KindOptions    api.KindOptions    `yaml:"kindOptions,omitempty"`
	Repo           string             `yaml:"repo,omitempty"`
	Timeout        int                `yaml:"timeout,omitempty"`

	// Root is a directory path relative to the parent directory of this
	// task definition which defines what directory should be included
	// in the task's Docker image.
	//
	// If not set, defaults to "." (in other words, the parent directory of this task definition).
	//
	// This field is ignored when using the pre-built image builder (aka "manual").
	Root string `yaml:"root,omitempty"`
}

func (d Definition_0_2) upgrade() (Definition, error) {
	return Definition(d), nil
}
