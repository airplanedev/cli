// fsx includes extensions to the stdlib fs package.
package fsx

import (
	"fmt"
	"os"
	"path"
)

// Exists returns true if the given path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FilesExist ensures that all paths exists or returns an error.
func AssertExistsAll(paths ...string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("build: the file %s is required", path.Base(p))
		}
	}
	return nil
}
