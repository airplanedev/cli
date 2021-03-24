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
