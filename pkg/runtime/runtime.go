// Package runtime generates code to match a runtime.
//
// The runtime package is capable of writing airplane specific
// comments that are used to link a task file to a remote task.
//
// All runtimes are also capable of generating initial code to
// match the task, including the parameters.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
)

var (
	// ErrMissing is returned when a resource was not found.
	//
	// It can be checked via `errors.Is(err, ErrMissing)`.
	ErrMissing = errors.New("runtime: resource is missing")

	// ErrNotImplemented is returned when a runtime does not
	// support preparing a run.
	//
	// It can be checked via `errors.Is(err, ErrNotImplemented)`.
	ErrNotImplemented = errors.New("runtime: not implemented")
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
	// Typically runtimes will look for a specific file such as
	// `package.json` or `requirements.txt`, they'll use `fs.Find()`.
	Root(path string) (dir string, err error)

	// Kind returns a task kind that matches the runtime.
	//
	// Generate and other methods should not be called
	// for a task that doesn't match the returned kind.
	Kind() api.TaskKind

	// FormatComment formats a string into a comment using
	// the relevant comment characters for this runtime.
	FormatComment(s string) string

	// PrepareRun should prepare a local run of a task.
	//
	// It must create a temporary directory, install any dependencies
	// and prepare the script to be run.
	//
	// On success the method returns a slice that represents an `cmd.Exec`
	// options which contains the command to be run and its arguments.
	//
	// If running the script locally is not supported the method returns
	// an `ErrNotImplemented`.
	PrepareRun(ctx context.Context, opts PrepareRunOptions) ([]string, error)
}

type PrepareRunOptions struct {
	// Path is the file path leading to the task's entrypoint.
	//
	// It should be an absolute path.
	Path string

	// ParamValues specifies the user-provided parameter values to
	// execute this run with.
	ParamValues api.Values

	// KindOptions specifies any runtime-specific task configuration.
	KindOptions api.KindOptions
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

// SuggestExt returns the default extension for a given TaskKind, if any.
func SuggestExt(kind api.TaskKind) string {
	for ext, runtime := range runtimes {
		if runtime.Kind() == kind {
			return ext
		}
	}
	return ""
}
