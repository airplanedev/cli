package build

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Golang creates a dockerfile for Go.
func golang(root string, args Args) (string, error) {
	var gomod = filepath.Join(root, "go.mod")
	var gosum = filepath.Join(root, "go.sum")
	var entrypoint = args["entrypoint"]
	var main = filepath.Join(root, entrypoint)

	if err := exist(gomod, gosum, main); err != nil {
		return "", err
	}

	t, err := template.New("golang").Parse(`
FROM golang:1.16.0-alpine3.13 as builder
WORKDIR /airplane
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENTRYPOINT ["go", "run", "/airplane/{{ .Main }}"]
`)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	var data struct {
		Root  string
		Main  string
		Gomod string
		Gosum string
	}
	data.Root = root
	data.Main = entrypoint

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Exist ensures that all paths exists or returns an error.
func exist(paths ...string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("build: the file %s is required", path.Base(p))
		}
	}
	return nil
}
