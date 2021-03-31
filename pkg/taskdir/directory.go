package taskdir

import (
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type TaskDirectory struct {
	// rootPath is the absolute path to the task's root directory.
	rootPath string
	// path is the absolute path of the airplane.yml task definition.
	defPath string
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
		td.defPath, td.closer, err = openGitHubDirectory(file)
		if err != nil {
			return TaskDirectory{}, err
		}
	} else {
		td.defPath, err = filepath.Abs(file)
		if err != nil {
			return TaskDirectory{}, errors.Wrap(err, "converting local file path to absolute path")
		}
	}

	def, err := td.ReadDefinition()
	if err != nil {
		return TaskDirectory{}, err
	}
	td.rootPath = path.Join(filepath.Dir(td.defPath), def.Root)

	if !strings.HasPrefix(td.defPath, td.rootPath+string(filepath.Separator)) {
		return TaskDirectory{}, errors.Errorf("%s must be inside of the task's root directory: %s", path.Base(td.defPath), td.rootPath)
	}

	return td, nil
}

func (this TaskDirectory) DefinitionPath() string {
	return this.defPath
}

func (this TaskDirectory) DefinitionRootPath() string {
	return this.rootPath
}

func (this TaskDirectory) Close() error {
	if this.closer != nil {
		return this.closer.Close()
	}

	return nil
}
