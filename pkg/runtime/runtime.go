// Package runtime generates code to match a runtime.
//
// The runtime package is capable of writing airplane specific
// comments that are used to link a task file to a remote task.
//
// All runtimes are also capable of generating initial code to
// match the task, including the parameters.
package runtime

import (
	"fmt"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
)

// Interface repersents a runtime.
type Interface interface {
	// Generate accepts a task and generates code to match the task.
	//
	// An error is returned if the code cannot be generated.
	Generate(task api.Task) ([]byte, error)

	// Slug returns the slug from the given code.
	//
	// If the comment was not found in code, the method
	// returns empty string and false.
	Slug(code []byte) (string, bool)

	// Comment returns a special airplane comment.
	//
	// The comment links a remote task to a file.
	Comment(task api.Task) string
}

// Runtimes is a collection of registered runtimes.
//
// The key is the file extension used for the runtime.
var runtimes = make(map[string]Interface)

// Register registers the given ext with r.
func Register(ext string, r Interface) {
	if _, ok := runtimes[ext]; ok {
		panic(fmt.Sprintf("runtime: %s already registered", ext))
	}
	runtimes[ext] = r
}

// Lookup returns a runtime by path.
func Lookup(path string) (Interface, bool) {
	ext := filepath.Ext(path)
	r, ok := runtimes[ext]
	return r, ok
}
