package taskdef

import (
	"io"
	"io/ioutil"
	"strings"

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
	var r io.Reader
	if strings.HasPrefix(path, "http://") {
		return Definition{}, errors.New("http:// paths are not supported, use https:// instead")
	} else if gitHubRegex.MatchString(path) {
		r = github{path}
	} else {
		r = file{path}
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return Definition{}, err
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
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "reading task definition")
	}

	// Find the first line without a document break. This is where
	// we'll insert the slug.
	lines := strings.Split(string(b), "\n")
	var idx int
	for idx = range lines {
		if !strings.HasPrefix(lines[idx], "---") {
			break
		}
	}

	contents := strings.Join([]string{
		strings.Join(lines[:idx], "\n"),
		"slug: " + slug,
		strings.Join(lines[idx:], "\n"),
	}, "\n")

	if err := ioutil.WriteFile(path, []byte(contents), 0); err != nil {
		return errors.Wrap(err, "updating task definition")
	}

	return nil
}
