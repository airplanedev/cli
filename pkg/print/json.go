package print

import (
	"encoding/json"
	"os"

	"github.com/airplanedev/cli/pkg/api"
)

// JSON implements a JSON formatter.
type JSON struct {
	encErr *json.Encoder
	encOut *json.Encoder
}

// NewJSONFormatter returns a new json formatter.
func NewJSONFormatter() *JSON {
	encErr := json.NewEncoder(os.Stderr)
	encErr.SetIndent("", "  ")

	encOut := json.NewEncoder(os.Stdout)
	return &JSON{
		encErr,
		encOut,
	}
}

// Tasks implementation.
func (j *JSON) tasks(tasks []api.Task) {
	j.encErr.Encode(tasks)
}

// Task implementation.
func (j *JSON) task(task api.Task) {
	j.encErr.Encode(task)
}

// Runs implementation.
func (j *JSON) runs(runs []api.Run) {
	j.encErr.Encode(runs)
}

// Run implementation.
func (j *JSON) run(run api.Run) {
	j.encErr.Encode(run)
}

func (j *JSON) outputs(outputs api.Outputs) {
	for key, values := range outputs {
		for _, value := range values {
			j.encOut.Encode(api.OutputRow{
				OutputName: key,
				Value:      value,
			})
		}
	}
}
