package taskdef

import (
	"io/ioutil"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Definition represents a YAML-based task definition that can be used to create
// or update Airplane tasks.
//
// Note this is the subset of fields that can be represented with a revision,
// and therefore isolated to a specific environment.
type Definition struct {
	Slug           string            `yaml:"slug"`
	Name           string            `yaml:"name"`
	Description    string            `yaml:"description"`
	Image          string            `yaml:"image"`
	Command        []string          `yaml:"command"`
	Arguments      []string          `yaml:"arguments"`
	Parameters     api.Parameters    `yaml:"parameters"`
	Constraints    api.Constraints   `yaml:"constraints"`
	Env            map[string]string `yaml:"env"`
	ResourceLimits map[string]string `yaml:"resourceLimits"`
	Builder        string            `yaml:"builder"`
	BuilderConfig  map[string]string `yaml:"builderConfig"`
	Repo           string            `yaml:"repo"`
	Timeout        int               `yaml:"timeout"`
}

func Read(path string) (Definition, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return Definition{}, errors.Wrapf(err, "reading task definition from %s", path)
	}

	var def Definition
	if err := yaml.Unmarshal(buf, &def); err != nil {
		return Definition{}, errors.Wrap(err, "unmarshaling task definition")
	}

	return def, nil
}
