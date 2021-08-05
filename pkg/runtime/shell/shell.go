package shell

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
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
	return nil, errors.New("not supported")
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
		logger.Warning("No Dockerfile file found - using a basic Ubuntu image.")
		logger.Log("To build a custom environment for your task, add a Dockerfile to the script directory (or any parent directory).")
		return filepath.Dir(path), nil
	}
	logger.Log("Using Dockerfile at %s to build the script environment", filepath.Join(root, "Dockerfile"))
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
