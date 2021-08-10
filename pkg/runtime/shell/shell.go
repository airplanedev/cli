package shell

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/aymerick/raymond"
	"github.com/pkg/errors"
)

// Init register the runtime.
func init() {
	runtime.Register(".sh", Runtime{})
}

// Code template.
var code = template.Must(template.New("sh").Parse(`#!/bin/bash
{{.Comment}}

# Params are in environment variables as AP_{SLUG}, e.g. AP_USER_ID
echo "Hello World!"
echo "Printing env for debugging purposes:"
env
`))

// Data represents the data template.
type data struct {
	Comment string
}

// Runtime implementation.
type Runtime struct{}

// PrepareRun implementation.
func (r Runtime) PrepareRun(ctx context.Context, opts runtime.PrepareRunOptions) ([]string, error) {
	root, err := r.Root(opts.Path)
	if err != nil {
		return nil, err
	}

	if dockerfilePath := build.FindDockerfile(root); dockerfilePath != "" {
		logger.Warning("Found Dockerfile at %s.", dockerfilePath)
		logger.Warning("`airplane dev` does not currently support running inside a Docker image.")
		logger.Warning("The script will run inside your local machine environment.")
	}

	if err := os.Mkdir(filepath.Join(root, ".airplane"), os.ModeDir|0777); err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "creating .airplane directory")
	}

	entrypoint, err := filepath.Rel(root, opts.Path)
	if err != nil {
		return nil, errors.Wrap(err, "entrypoint is not within the task root")
	}
	shim, err := build.ShellShim(entrypoint)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, ".airplane/shim.sh"), []byte(shim), 0644); err != nil {
		return nil, errors.Wrap(err, "writing shim file")
	}

	cmd := []string{"bash", filepath.Join(root, ".airplane/shim.sh")}
	// TODO: this is a rough approximation of how interpolateParameters works in prod
	for slug, _ := range opts.ParamValues {
		tmpl := fmt.Sprintf("%s={{%s}}", slug, slug)
		val, err := raymond.Render(tmpl, opts.ParamValues)
		if err != nil {
			return nil, errors.Wrap(err, "rendering shell command")
		}
		cmd = append(cmd, val)
	}
	return cmd, nil
}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{Comment: runtime.Comment(r, t)}
	var buf bytes.Buffer

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("shell: template execute - %w", err)
	}

	return buf.Bytes(), nil
}

// Workdir implementation.
func (r Runtime) Workdir(path string) (string, error) {
	return r.Root(path)
}

// Root implementation.
func (r Runtime) Root(path string) (string, error) {
	root, ok := fsx.Find(path, "Dockerfile")
	if !ok {
		return filepath.Dir(path), nil
	}
	return root, nil
}

// Kind implementation.
func (r Runtime) Kind() api.TaskKind {
	return api.TaskKindShell
}

// FormatComment implementation.
func (r Runtime) FormatComment(s string) string {
	var lines []string

	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, "# "+line)
	}

	return strings.Join(lines, "\n")
}
