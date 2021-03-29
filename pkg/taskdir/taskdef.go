package taskdir

import (
	"io/ioutil"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Definition represents a YAML-based task definition that can be used to create
// or update Airplane tasks.
//
// Note this is the subset of fields that can be represented with a revision,
// and therefore isolated to a specific environment.
type Definition struct {
	Slug           string             `yaml:"slug"`
	Name           string             `yaml:"name"`
	Description    string             `yaml:"description,omitempty"`
	Image          string             `yaml:"image,omitempty"`
	Command        []string           `yaml:"command,omitempty"`
	Arguments      []string           `yaml:"arguments,omitempty"`
	Parameters     api.Parameters     `yaml:"parameters,omitempty"`
	Constraints    api.Constraints    `yaml:"constraints,omitempty"`
	Env            map[string]string  `yaml:"env,omitempty"`
	ResourceLimits api.ResourceLimits `yaml:"resourceLimits,omitempty"`
	Builder        string             `yaml:"builder,omitempty"`
	BuilderConfig  api.BuilderConfig  `yaml:"builderConfig,omitempty"`
	Repo           string             `yaml:"repo,omitempty"`
	Timeout        int                `yaml:"timeout,omitempty"`
}

func (this Definition) Validate() (Definition, error) {
	canPrompt := utils.CanPrompt()

	if this.Slug == "" {
		if !canPrompt {
			return this, errors.New("Expected a slug")
		}

		if err := survey.AskOne(
			&survey.Input{
				Message: "Pick a unique identifier (slug) for this task",
				Default: utils.MakeSlug(this.Name),
			},
			&this.Slug,
			survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
			survey.WithValidator(func(val interface{}) error {
				if str, ok := val.(string); !ok || !utils.IsSlug(str) {
					return errors.New("Slugs can only contain lowercase letters, underscores, and numbers.")
				}

				return nil
			}),
		); err != nil {
			return this, errors.Wrap(err, "prompting for slug")
		}
	}

	// TODO: persist validation changes, if any, back to the local file.
	// TODO: validate the rest of the fields!

	return this, nil
}

func (this TaskDirectory) DefinitionPath() string {
	return this.path
}

func (this TaskDirectory) ReadDefinition() (Definition, error) {
	buf, err := ioutil.ReadFile(this.path)
	if err != nil {
		return Definition{}, errors.Wrap(err, "reading task definition")
	}

	var def Definition
	if err := yaml.Unmarshal(buf, &def); err != nil {
		return Definition{}, errors.Wrap(err, "unmarshaling task definition")
	}

	return def, nil
}

func (this TaskDirectory) WriteDefinition(def Definition) error {
	data, err := yaml.Marshal(def)
	if err != nil {
		return errors.Wrap(err, "marshalling definition")
	}

	if err := ioutil.WriteFile(this.path, data, 0664); err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}
