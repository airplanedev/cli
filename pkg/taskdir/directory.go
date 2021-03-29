package taskdir

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type TaskDirectory struct {
	// Dir is the absolute local path to the directory that TaskDirectory represents.
	Dir string

	// path is the local path, relative to Dir, of the airplane.yml task definition.
	path string
	// closer is used to clean up TaskDirectory.
	closer io.Closer
}

func Open(file string) (TaskDirectory, error) {
	if strings.HasPrefix(file, "http://") {
		return TaskDirectory{}, errors.New("http:// paths are not supported, use https:// instead")
	}

	var td TaskDirectory
	var err error
	if strings.HasPrefix(file, "github.com/") || strings.HasPrefix(file, "https://github.com/") {
		td.path, td.closer, err = openGitHubDirectory(file)
		if err != nil {
			return TaskDirectory{}, err
		}
	} else {
		td.path = file
	}

	td.Dir, err = filepath.Abs(filepath.Dir(td.path))
	if err != nil {
		return TaskDirectory{}, errors.Wrap(err, "parsing file directory")
	}

	return td, nil
}

func (this TaskDirectory) Close() error {
	if this.closer != nil {
		return this.closer.Close()
	}

	return nil
}
