// Package runtime generates code to match a runtime.
//
// The runtime package is capable of writing airplane specific
// comments that are used to link a task file to a remote task.
//
// All runtimes are also capable of generating initial code to
// match the task, including the parameters.
package runtime

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fs"
)

var (
	// ErrMissing is returned when a resource was not found.
	//
	// It can be checked via `errors.Is(err, ErrMissing)`.
	ErrMissing = errors.New("runtime: resource is missing")
)

// Settings represent Airplane specific settings.
type Settings struct {
	Root string `json:"root"`
}

// Interface repersents a runtime.
type Interface interface {
	// Generate accepts a task and generates code to match the task.
	//
	// An error is returned if the code cannot be generated.
	Generate(task api.Task) ([]byte, error)

	// Workdir attempts to detect the root of the given task path.
	//
	// Unlike root it decides the dockerfile's `workdir` directive
	// this might be different than root because it decides where
	// the build commands are run.
	Workdir(path string) (dir string, err error)

	// Root attempts to detect the root of the given task path.
	//
	// It returns the suggested root, if a root directory is not
	// found the method returns an `ErrMissing`.
	//
	// Typically runtimes will look for a specific file such as
	// `package.json` or `requirements.txt`, they'll use `runtime.Pathof()`.
	Root(path string) (dir string, err error)

	// Kind returns a task kind that matches the runtime.
	//
	// Generate and other methods should not be called
	// for a task that doesn't match the returned kind.
	Kind() api.TaskKind
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

const (
	// Separator converted to string to abort pathof at root.
	sep = string(filepath.Separator)
)

// Pathof attempts to find the path of the given filename.
//
// The method recursively visits parent dirs until the given
// filename is found, If the file is not found the method
// returns an `ErrMissing`.
func Pathof(parent, filename string) (string, error) {
	dst := filepath.Join(parent, filename)

	if !fs.Exists(dst) {
		next := filepath.Dir(parent)
		if next == "." || next == sep {
			return "", ErrMissing
		}
		return Pathof(next, filename)
	}

	return parent, nil
}
