package python

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/pkg/errors"
)

// Init register the runtime.
func init() {
	runtime.Register(".py", Runtime{})
}

// Code template.
var code = template.Must(template.New("py").Parse(`{{.Comment}}

def main(params):
    print("parameters:", params)
`))

// Data represents the data template.
type data struct {
	Comment string
}

// Runtime implementation.
type Runtime struct{}

// PrepareRun implementation.
func (r Runtime) PrepareRun(ctx context.Context, opts runtime.PrepareRunOptions) ([]string, error) {
	if err := checkPythonInstalled(ctx); err != nil {
		return nil, err
	}

	root, err := r.Root(opts.Path)
	if err != nil {
		return nil, err
	}

	if err := os.Mkdir(filepath.Join(root, ".airplane"), os.ModeDir|0777); err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "creating .airplane directory")
	}

	entrypoint, err := filepath.Rel(root, opts.Path)
	if err != nil {
		return nil, errors.Wrap(err, "entrypoint is not within the task root")
	}
	shim, err := build.PythonShim(root, entrypoint)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, ".airplane/shim.py"), []byte(shim), 0644); err != nil {
		return nil, errors.Wrap(err, "writing shim file")
	}

	pv, err := json.Marshal(opts.ParamValues)
	if err != nil {
		return nil, errors.Wrap(err, "serializing param values")
	}

	return []string{"python3", filepath.Join(root, ".airplane/shim.py"), string(pv)}, nil
}

// Checks for python3 binary, as per PEP 0394:
// https://www.python.org/dev/peps/pep-0394/#recommendation
func checkPythonInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "python3", "--version")
	logger.Debug("Running %s", logger.Bold(strings.Join(cmd.Args, " ")))
	err := cmd.Run()
	if err != nil {
		return errors.New(heredoc.Doc(`
		It looks like the python3 command is not installed.

		Ensure Python 3 is installed and the python3 command exists: https://www.python.org/downloads
	`))
	}
	return nil
}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{Comment: runtime.Comment(r, t)}
	var buf bytes.Buffer

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("python: template execute - %w", err)
	}

	return buf.Bytes(), nil
}

// Workdir implementation.
func (r Runtime) Workdir(path string) (string, error) {
	return r.Root(path)
}

// Root implementation.
func (r Runtime) Root(path string) (string, error) {
	root, ok := fsx.Find(path, "requirements.txt")
	if !ok {
		return filepath.Dir(path), nil
	}
	return root, nil
}

// Kind implementation.
func (r Runtime) Kind() api.TaskKind {
	return api.TaskKindPython
}

// FormatComment implementation.
func (r Runtime) FormatComment(s string) string {
	var lines []string

	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, "# "+line)
	}

	return strings.Join(lines, "\n")
}
