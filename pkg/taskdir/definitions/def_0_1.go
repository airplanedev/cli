package definitions

import (
	"github.com/airplanedev/cli/pkg/api"
)

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
	return Definition_0_2{
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
	}.upgrade()
}
