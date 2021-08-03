package build

import (
	_ "embed"
	"io/ioutil"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

func shell(root string, options api.KindOptions) (string, error) {
	// Assert that the entrypoint file exists:
	entrypoint, _ := options["entrypoint"].(string)
	if err := fsx.AssertExistsAll(filepath.Join(root, entrypoint)); err != nil {
		return "", err
	}

	shim, err := ShellShim(entrypoint)
	if err != nil {
		return "", err
	}

	// Build off of the dockerfile if provided:
	var baseDockerfile string
	dockerfile, _ := options["dockerfile"].(string)
	if dockerfile == "" {
		baseDockerfile = heredoc.Doc(`
			FROM ubuntu:21.04
			# TODO: install some basics
		`)
	} else {
		dockerfilePath := filepath.Join(root, dockerfile)
		if err := fsx.AssertExistsAll(dockerfilePath); err != nil {
			return "", err
		}
		contents, err := ioutil.ReadFile(dockerfilePath)
		if err != nil {
			return "", errors.Wrap(err, "opening dockerfile")
		}
		baseDockerfile = string(contents)
	}

	dfTemplate := baseDockerfile + heredoc.Doc(`

		WORKDIR /airplane
		RUN mkdir -p .airplane && {{.InlineShim}} > .airplane/shim.sh
		
		COPY . .
		RUN chmod +x {{.Entrypoint}}
		
		ENTRYPOINT ["bash", ".airplane/shim.sh"]
	`)
	df, err := applyTemplate(dfTemplate, struct {
		InlineShim string
		Entrypoint string
	}{
		InlineShim: inlineString(shim),
		Entrypoint: entrypoint,
	})
	if err != nil {
		return "", err
	}

	return df, nil
}

//go:embed shell-shim.sh
var shellShim string

func ShellShim(entrypoint string) (string, error) {
	// exec needs a relative path
	entrypoint = "./" + filepath.Clean(entrypoint)
	shim, err := applyTemplate(shellShim, struct {
		Entrypoint string
	}{
		Entrypoint: entrypoint,
	})
	if err != nil {
		return "", errors.Wrap(err, "templating shim")
	}

	return shim, nil
}
