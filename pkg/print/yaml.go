package print

import (
	"os"

	"github.com/airplanedev/cli/pkg/api"
	"gopkg.in/yaml.v2"
)

// YAML implements a YAML formatter.
//
// Its zero-value is ready for use.
type YAML struct{}

// Tasks implementation.
func (YAML) tasks(tasks []api.Task) {
	yaml.NewEncoder(os.Stderr).Encode(tasks)
}

// Task implementation.
func (YAML) task(task api.Task) {
	yaml.NewEncoder(os.Stderr).Encode(task)
}
