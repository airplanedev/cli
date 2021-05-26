package fs

import "os"

// Exists returns true if the given path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
