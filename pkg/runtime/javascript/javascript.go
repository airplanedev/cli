package javascript

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
)

// Init register the runtime.
func init() {
	runtime.Register(".js", Runtime{})
}

// Code template.
var code = template.Must(template.New("js").Parse(`{{.Comment}}

export default async function(params) {
  console.log('parameters:', params);
}
`))

// Data represents the data template.
type data struct {
	Comment string
}

// Runtime implementaton.
type Runtime struct{}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{Comment: runtime.Comment(r, t)}
	var buf bytes.Buffer

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("javascript: template execute - %w", err)
	}

	return buf.Bytes(), nil
}

// Workdir implementation.
func (r Runtime) Workdir(path string) (string, error) {
	return runtime.Pathof(path, "package.json")
}

// Root implementation.
//
// The method finds the nearest package.json, If the package.json contains
// any airplane settings with `root` definition it will use that as the root.
func (r Runtime) Root(path string) (string, error) {
	dst, err := runtime.Pathof(path, "package.json")
	if err != nil {
		return "", err
	}

	pkgjson := filepath.Join(dst, "package.json")
	buf, err := ioutil.ReadFile(pkgjson)
	if err != nil {
		return "", errors.Wrapf(err, "javascript: reading %s", dst)
	}

	var pkg struct {
		Settings runtime.Settings `json:"airplane"`
	}

	if err := json.Unmarshal(buf, &pkg); err != nil {
		return "", fmt.Errorf("javascript: reading %s - %w", dst, err)
	}

	if root := pkg.Settings.Root; root != "" {
		return filepath.Join(dst, root), nil
	}

	return dst, nil
}

// Kind implementation.
func (r Runtime) Kind() api.TaskKind {
	return api.TaskKindNode
}

func (r Runtime) FormatComment(s string) string {
	lines := []string{}
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, "// "+line)
	}
	return strings.Join(lines, "\n")
}

func (r Runtime) PrepareRun(ctx context.Context, path string, paramValues api.Values) ([]string, error) {
	root, err := r.Root(path)
	if err != nil {
		return nil, err
	}

	if err := os.Mkdir(filepath.Join(root, ".airplane"), os.ModeDir|0777); err != nil {
		if !strings.HasSuffix(err.Error(), "file exists") {
			return nil, errors.Wrap(err, "creating .airplane directory")
		}
	}

	shim, err := utils.ApplyTemplate(build.NodeShim, struct {
		ImportPath string
	}{
		ImportPath: "main.js", //relimport
	})
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, ".airplane/shim.ts"), []byte(shim), 0644); err != nil {
		return nil, errors.Wrap(err, "writing shim file")
	}

	if err := os.RemoveAll(filepath.Join(root, ".airplane/dist")); err != nil {
		return nil, errors.Wrap(err, "cleaning dist folder")
	}

	if utils.FilesExist(filepath.Join(root, "package.json")) != nil {
		if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0777); err != nil {
			return nil, errors.Wrap(err, "creating default package.json")
		}
	}

	isYarn := utils.FilesExist(filepath.Join(root, "yarn.lock")) == nil
	var cmd *exec.Cmd
	if isYarn {
		cmd = exec.CommandContext(ctx, "yarn", "add", "-D", "@types/node")
	} else {
		cmd = exec.CommandContext(ctx, "npm", "install", "--save-dev", "@types/node")
	}
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		return nil, errors.New("failed to add @types/node dependency")
	}

	// TODO: warn if Node major version does not match
	// TODO: install tsc
	// TODO: es2019 if nodeVersion
	// TODO: support root vs. workdir

	cmd = exec.CommandContext(ctx,
		"tsc",
		"--allowJs",
		"--module", "commonjs",
		"--target", "es2020",
		"--lib", "es2020",
		"--esModuleInterop",
		"--outDir", ".airplane/dist",
		"--rootDir", ".",
		"--skipLibCheck",
		"--pretty",
		".airplane/shim.ts")
	cmd.Dir = root // workdir?
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log(strings.TrimSpace(string(out)))
		logger.Debug("\nCommand: %s", strings.Join(cmd.Args, " "))

		return nil, errors.Errorf("failed to compile %s", path)
	}

	pv, err := json.Marshal(paramValues)
	if err != nil {
		return nil, errors.Wrap(err, "serializing param values")
	}

	return []string{"node", filepath.Join(root, ".airplane/dist/.airplane/shim.js"), string(pv)}, nil
}
