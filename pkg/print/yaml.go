package print

import (
	"errors"
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

// Runs implementation.
func (YAML) runs(runs []api.Run) {
	yaml.NewEncoder(os.Stderr).Encode(runs)
}

// Run implementation.
func (YAML) run(run api.Run) {
	yaml.NewEncoder(os.Stderr).Encode(run)
}

func (YAML) outputs(outputs api.Outputs) {
	errors.New("Not implemented")
}
