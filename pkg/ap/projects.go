package ap

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

const (
	projectfile = "airplane.json"
)

var (
	// ErrNoProject represents a missing project file.
	ErrNoProject = errors.New("ap: missing " + projectfile)
)

// Project represents an Airplane project.
//
// On the filesystem it is persisted as `airplane.json`
// at the root of the project.
type Project struct {
	root      string
	Framework string   `json:"framework"`
	Tasks     []string `json:"tasks"`
}

// NewProject returns a new empty project at root.
func NewProject(root string) *Project {
	return &Project{root: root}
}

// ReadProject reads a project from the given root.
func ReadProject(root string) (*Project, error) {
	var path = filepath.Join(root, projectfile)
	var proj = Project{root: root}

	if !fsx.Exists(path) {
		return nil, ErrNoProject
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", projectfile)
	}

	if err := json.Unmarshal(buf, &proj); err != nil {
		return nil, errors.Wrapf(err, "unmarshal %s", projectfile)
	}

	return &proj, nil
}

// Save saves the project.
func (p *Project) Save() error {
	var path = filepath.Join(p.root, projectfile)

	buf, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal project")
	}

	err = ioutil.WriteFile(path, buf, 0600)
	if err != nil {
		return errors.Wrapf(err, "write %s", projectfile)
	}

	return nil
}
