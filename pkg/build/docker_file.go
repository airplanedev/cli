package build

import (
	"io/ioutil"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

func dockerfile(root string, buildConfig api.BuildConfig, options api.KindOptions) (string, error) {
	var dockerfile string
	if buildConfig.Kind != "" {
		dockerfile = buildConfig.Dockerfile
	} else {
		dockerfile, _ = options["dockerfile"].(string)
	}
	dockerfilePath := filepath.Join(root, dockerfile)
	if err := fsx.AssertExistsAll(dockerfilePath); err != nil {
		return "", err
	}

	contents, err := ioutil.ReadFile(dockerfilePath)
	if err != nil {
		return "", errors.Wrap(err, "opening dockerfile")
	}

	return string(contents), nil
}
