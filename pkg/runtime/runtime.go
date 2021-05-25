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
	"os"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fs"
)

// Interface repersents a runtime.
type Interface interface {
	// Generate accepts a task and generates code to match the task.
	//
	// An error is returned if the code cannot be generated.
	Generate(task api.Task) ([]byte, error)
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

// Lookup returns a runtikme by ext.
func Lookup(ext string) (Interface, bool) {
	r, ok := runtimes[ext]
	return r, ok
}

// Code returns the code for path.
func Code(path string, t api.Task) ([]byte, error) {
	var ext = filepath.Ext(path)

	r, ok := Lookup(ext)
	if !ok {
		return nil, fmt.Errorf("cannot scaffold for extension %s", ext)
	}

	code, err := r.Generate(t)
	if err != nil {
		return nil, err
	}

	return code, nil
}

// Generate attempts to generate code for task and writes it to path.
func Generate(path string, t api.Task) error {
	if fs.Exists(path) {
		return fmt.Errorf("path %s already exists", path)
	}

	code, err := Code(path, t)
	if err != nil {
		return err
	}

	return os.WriteFile(path, code, 0644)
}
