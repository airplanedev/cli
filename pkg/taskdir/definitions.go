package taskdir

import (
	"github.com/airplanedev/cli/pkg/api"
	"github.com/pkg/errors"
)

// Definition represents a YAML-based task definition that can be used to create
// or update Airplane tasks.
//
// Note this is the subset of fields that can be represented with a revision,
// and therefore isolated to a specific environment.
type Definition Definition_0_2

func (this Definition) Validate() (Definition, error) {
	if this.Slug == "" {
		return this, errors.New("Expected a task slug")
	}

	// TODO: validate the rest of the fields!

	return this, nil
}

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

type Definition_0_1 struct {
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
	Builder        string             `yaml:"builder,omitempty"`
	BuilderConfig  api.KindOptions    `yaml:"builderConfig,omitempty"`
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

func (d Definition_0_1) upgrade() (Definition, error) {
	return Definition{
		Slug:           d.Slug,
		Name:           d.Name,
		Description:    d.Description,
		Image:          d.Image,
		Command:        d.Command,
		Arguments:      d.Arguments,
		Parameters:     d.Parameters,
		Constraints:    d.Constraints,
		Env:            d.Env,
		ResourceLimits: d.ResourceLimits,
		Kind:           d.Builder,
		KindOptions:    d.BuilderConfig,
		Repo:           d.Repo,
		Timeout:        d.Timeout,
		Root:           d.Root,
	}, nil
}
