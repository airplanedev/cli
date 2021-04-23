package taskdir

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const taskDefDocURL = "https://docs.airplane.dev/reference/task-definition-reference"

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
	Constraints    api.RunConstraints `yaml:"constraints,omitempty"`
	Env            api.TaskEnv        `yaml:"env,omitempty"`
	ResourceLimits api.ResourceLimits `yaml:"resourceLimits,omitempty"`
	Builder        string             `yaml:"builder,omitempty"`
	BuilderConfig  api.BuilderConfig  `yaml:"builderConfig,omitempty"`
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

func (this Definition) Validate() (Definition, error) {
	if this.Slug == "" {
		return this, errors.New("Expected a task slug")
	}

	// TODO: validate the rest of the fields!

	return this, nil
}

func (this TaskDirectory) ReadDefinition() (Definition, error) {
	buf, err := ioutil.ReadFile(this.defPath)
	if err != nil {
		return Definition{}, errors.Wrap(err, "reading task definition")
	}

	// Validate definition against our Definition struct
	if err := ValidateYAML(buf, Definition{}); err != nil {
		defPath := this.defPath
		// Attempt to set a prettier defPath, best effort
		if wd, err := os.Getwd(); err != nil {
			logger.Debug("%s", err)
		} else if path, err := filepath.Rel(wd, defPath); err != nil {
			logger.Debug("%s", err)
		} else {
			defPath = path
		}
		// Print any "expected" validation errors
		switch err := errors.Cause(err).(type) {
		case ErrInvalidYAML:
			logger.Log(logger.Red("\nError reading %s: invalid YAML", defPath))
			logger.Log("\nTask definition reference: %s", taskDefDocURL)
		case ErrSchemaValidation:
			logger.Log(logger.Red("\nError reading %s:\n", defPath))
			for _, verr := range err.Errors {
				logger.Log("  %s: %s", verr.Field(), verr.Description())
			}
			logger.Log("\nTask definition reference: %s", taskDefDocURL)
		}
		return Definition{}, errors.Wrapf(err, "error reading %s", defPath)
	}

	var def Definition
	if err := yaml.Unmarshal(buf, &def); err != nil {
		return Definition{}, errors.Wrap(err, "unmarshalling task definition")
	}

	return def, nil
}

// WriteSlug updates the slug of a task definition and persists this to disk.
//
// It attempts to retain the existing file's formatting (comments, etc.) where possible.
func (this TaskDirectory) WriteSlug(slug string) error {
	if err := utils.SetYAMLField(this.defPath, "slug", slug); err != nil {
		return errors.Wrap(err, "setting slug")
	}

	return nil
}

func (this TaskDirectory) WriteDefinition(def Definition) error {
	data, err := yaml.Marshal(def)
	if err != nil {
		return errors.Wrap(err, "marshalling definition")
	}

	if err := ioutil.WriteFile(this.defPath, data, 0664); err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}
