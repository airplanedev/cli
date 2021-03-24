package taskdef

import (
	"io/ioutil"
	"os"

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

// WriteSlug inserts a task definition into a file and attempts to
// preserve the files existing format as much as possible.
func WriteSlug(path, slug string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return errors.Wrap(err, "opening task definition")
	}
	defer f.Close()

	node := yaml.Node{}
	if err := yaml.NewDecoder(f).Decode(&node); err != nil {
		return errors.Wrap(err, "unmarshaling task definition")
	}

	for _, subnode := range node.Content {
		// Find the root map, where we'll insert the slug.
		if subnode.Kind == yaml.MappingNode {
			subnode.Content = append([]*yaml.Node{
				&yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: "slug",
				},
				&yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: slug,
				},
			}, subnode.Content...)
		}
	}

	if _, err := f.Seek(0, 0); err != nil {
		return errors.Wrap(err, "seeking to start of task definition")
	}
	if err := f.Truncate(0); err != nil {
		return errors.Wrap(err, "truncating file")
	}
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(&node); err != nil {
		return errors.Wrap(err, "marshaling task definition")
	}

	return nil
}
