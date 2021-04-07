package print

import (
	"os"

	"github.com/airplanedev/cli/pkg/api"
	"gopkg.in/yaml.v3"
)

// YAML implements a YAML formatter.
//
// Its zero-value is ready for use.
type YAML struct{}

// Tasks implementation.
func (YAML) tasks(tasks []api.Task) {
	yaml.NewEncoder(os.Stdout).Encode(tasks)
}

// Task implementation.
func (YAML) task(task api.Task) {
	yaml.NewEncoder(os.Stdout).Encode(task)
}

// Runs implementation.
func (YAML) runs(runs []api.Run) {
	yaml.NewEncoder(os.Stdout).Encode(runs)
}

// Run implementation.
func (YAML) run(run api.Run) {
	yaml.NewEncoder(os.Stdout).Encode(run)
}

// Outputs implementation.
func (YAML) outputs(outputs api.Outputs) {
	var rows []api.OutputRow
	for key, values := range outputs {
		for _, value := range values {
			rows = append(rows, api.OutputRow{
				OutputName: key,
				Value:      value,
			})
		}
	}
	yaml.NewEncoder(os.Stdout).Encode(rows)
}

// Config implementation.
func (YAML) config(config api.Config) {
	yaml.NewEncoder(os.Stdout).Encode(config)
}
