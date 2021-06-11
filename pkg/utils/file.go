package utils

import (
	"fmt"
	"os"
	"path"
)

// exist ensures that all paths exists or returns an error.
func FilesExist(paths ...string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("build: the file %s is required", path.Base(p))
		}
	}
	return nil
}
