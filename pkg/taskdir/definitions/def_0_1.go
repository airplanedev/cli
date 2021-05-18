package definitions

import (
	"github.com/airplanedev/cli/pkg/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
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
	ResourceLimits map[string]string  `yaml:"resourceLimits,omitempty"`
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
	def := Definition_0_2{
		Slug:        d.Slug,
		Name:        d.Name,
		Description: d.Description,
		Arguments:   d.Arguments,
		Parameters:  d.Parameters,
		Constraints: d.Constraints,
		Env:         d.Env,
		Repo:        d.Repo,
		Timeout:     d.Timeout,
		Root:        d.Root,
	}

	if d.Builder == "deno" {
		def.Deno = &DenoDefinition{}
		if err := mapstructure.Decode(d.BuilderConfig, &def.Deno); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Deno options")
		}

	} else if d.Builder == "docker" {
		def.Dockerfile = &DockerDefinition{}
		if err := mapstructure.Decode(d.BuilderConfig, &def.Dockerfile); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Dockerfile options")
		}

	} else if d.Builder == "go" {
		def.Go = &GoDefinition{}
		if err := mapstructure.Decode(d.BuilderConfig, &def.Go); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Go options")
		}

	} else if d.Builder == "node" {
		def.Node = &NodeDefinition{}
		if err := mapstructure.Decode(d.BuilderConfig, &def.Node); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Node options")
		}

	} else if d.Builder == "python" {
		def.Python = &PythonDefinition{}
		if err := mapstructure.Decode(d.BuilderConfig, &def.Python); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Python options")
		}

	} else if d.Builder == "" {
		def.Manual = &ManualDefinition{
			Image:   d.Image,
			Command: d.Command,
		}
	}

	return def.upgrade()
}
