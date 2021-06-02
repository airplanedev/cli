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
	"github.com/airplanedev/cli/pkg/fs"
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

	// Root attempts to detect the root of the given task path.
	//
	// It returns the suggested root and `ok=true` if a suggestion
	// was found otherwise it returns an empty string.
	//
	// Typically runtimes will look for a specific file such as
	// `package.json` or `requirements.txt`, they'll use `runtime.Pathof()`.
	Root(path string) (dir string, ok bool)
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
// filename is found, ok reports if the filename is found
// and the string is the path.
func Pathof(parent, filename string) (string, bool) {
	dst := filepath.Join(parent, filename)

	if !fs.Exists(dst) {
		if parent == sep {
			return "", false
		}
		next := filepath.Dir(parent)
		return Pathof(next, filename)
	}

	return parent, true
}
