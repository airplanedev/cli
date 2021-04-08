package print

import (
	"encoding/json"
	"os"

	"github.com/airplanedev/cli/pkg/api"
)

// JSON implements a JSON formatter.
type JSON struct {
	enc *json.Encoder
}

// NewJSONFormatter returns a new json formatter.
func NewJSONFormatter() *JSON {
	return &JSON{
		enc: json.NewEncoder(os.Stdout),
	}
}

// APIKeys implementation.
func (j *JSON) apiKeys(apiKeys []api.APIKey) {
	j.enc.Encode(apiKeys)
}

// Tasks implementation.
func (j *JSON) tasks(tasks []api.Task) {
	j.enc.Encode(tasks)
}

// Task implementation.
func (j *JSON) task(task api.Task) {
	j.enc.Encode(task)
}

// Runs implementation.
func (j *JSON) runs(runs []api.Run) {
	j.enc.Encode(runs)
}

// Run implementation.
func (j *JSON) run(run api.Run) {
	j.enc.Encode(run)
}

// Outputs implementation.
func (j *JSON) outputs(outputs api.Outputs) {
	for key, values := range outputs {
		for _, value := range values {
			j.enc.Encode(api.OutputRow{
				OutputName: key,
				Value:      value,
			})
		}
	}
}

// Config implementation.
func (j *JSON) config(config api.Config) {
	j.enc.Encode(config)
}
