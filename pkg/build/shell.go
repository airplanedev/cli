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
	if entrypoint == "" {
		return "", errors.New("entrypoint is unexpectedly missing")
	}
	if err := fsx.AssertExistsAll(filepath.Join(root, entrypoint)); err != nil {
		return "", err
	}

	shim, err := ShellShim(entrypoint)
	if err != nil {
		return "", err
	}

	// Build off of the dockerfile if provided:
	var dockerfileTemplate string
	if dockerfilePath := FindDockerfile(root); dockerfilePath != "" {
		contents, err := ioutil.ReadFile(dockerfilePath)
		if err != nil {
			return "", errors.Wrap(err, "opening dockerfile")
		}
		dockerfileTemplate = string(contents)
	} else {
		dockerfileTemplate = heredoc.Doc(`
			FROM ubuntu:21.04
			# Install some common libraries
			RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
				&& apt-get -y install --no-install-recommends \
					apt-utils \
					openssh-client \
					gnupg2 \
					iproute2 \
					procps \
					lsof \
					htop \
					net-tools \
					curl \
					wget \
					ca-certificates \
					unzip \
					zip \
					nano \
					vim-tiny \
					less \
					jq \
					lsb-release \
					apt-transport-https \
					dialog \
					zlib1g \
					locales \
					strace \
				&& apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/*
		`)
	}

	// Extend template with our own logic - set up a WORKDIR and shim.
	dockerfileTemplate = dockerfileTemplate + heredoc.Doc(`
		WORKDIR /airplane
		RUN mkdir -p .airplane && {{.InlineShim}} > .airplane/shim.sh
		
		COPY . .
		RUN chmod +x {{.Entrypoint}}
		
		ENTRYPOINT ["bash", ".airplane/shim.sh"]
	`)
	return applyTemplate(dockerfileTemplate, struct {
		InlineShim string
		Entrypoint string
	}{
		InlineShim: inlineString(shim),
		Entrypoint: backslashEscape(entrypoint, `"`),
	})
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

// FindDockerfile looks for variants of supported Dockerfile locations and returns non-empty path
// to the file, if found.
func FindDockerfile(root string) string {
	for _, filePath := range []string{
		"Dockerfile.airplane",
		"Dockerfile",
	} {
		dockerfilePath := filepath.Join(root, filePath)
		if fsx.Exists(dockerfilePath) {
			return dockerfilePath
		}
	}
	return ""
}
