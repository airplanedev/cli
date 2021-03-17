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
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	return &JSON{
		enc: enc,
	}
}

// Tasks implementation.
func (j *JSON) tasks(tasks []api.Task) {
	j.enc.Encode(tasks)
}

// Task implementation.
func (j *JSON) task(task api.Task) {
	j.enc.Encode(task)
}
