package print

import (
	"encoding/json"
	"os"

	"github.com/airplanedev/cli/pkg/api"
)

// JSON implements a JSON formatter.
//
// Its zero-value is ready for use.
type JSON struct{}

// Tasks implementation.
func (JSON) tasks(tasks []api.Task) {
	json.NewEncoder(os.Stderr).Encode(tasks)
}
