// Package scaffold generates code to match a task.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fs"
)

// Interface repersents a generator.
type Interface interface {
	// Generate accepts a task and generates code to match the task.
	//
	// An error is returned if the code cannot be generated.
	Generate(task api.Task) ([]byte, error)
}

// Languages is a collection of registered languages.
//
// The key is the file extension used for the language.
var languages = make(map[string]Interface)

// Register registers the given ext with gen.
func Register(ext string, gen Interface) {
	if _, ok := languages[ext]; ok {
		panic(fmt.Sprintf("scaffold: %s already registered", ext))
	}
	languages[ext] = gen
}

// Lookup returns a generator by ext.
func Lookup(ext string) (Interface, bool) {
	gen, ok := languages[ext]
	return gen, ok
}

// Code returns the code for path.
func Code(path string, t api.Task) ([]byte, error) {
	var ext = filepath.Ext(path)

	g, ok := Lookup(ext)
	if !ok {
		return nil, fmt.Errorf("cannot scaffold for extension %s", ext)
	}

	code, err := g.Generate(t)
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
