package build

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

func dockerfile(root string, args Args) (string, error) {
	dockerfilePath := filepath.Join(root, args["dockerfile"])
	if err := exist(dockerfilePath); err != nil {
		return "", err
	}

	contents, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return "", errors.Wrap(err, "opening dockerfile")
	}

	return string(contents), nil
}
