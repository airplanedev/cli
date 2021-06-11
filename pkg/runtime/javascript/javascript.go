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

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/blang/semver/v4"
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

func (r Runtime) PrepareRun(ctx context.Context, opts runtime.PrepareRunOptions) ([]string, error) {
	checkNodeVersion(ctx, opts.KindOptions)
	if err := checkTscInstalled(ctx); err != nil {
		return nil, err
	}

	root, err := r.Root(opts.Path)
	if err != nil {
		return nil, err
	}
	workdir := filepath.Dir(opts.Path)

	if err := os.Mkdir(filepath.Join(root, ".airplane"), os.ModeDir|0777); err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "creating .airplane directory")
	}

	shim, err := build.NodeShim(root, opts.Path)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, ".airplane/shim.ts"), []byte(shim), 0644); err != nil {
		return nil, errors.Wrap(err, "writing shim file")
	}

	if err := os.RemoveAll(filepath.Join(root, ".airplane/dist")); err != nil {
		return nil, errors.Wrap(err, "cleaning dist folder")
	}

	if fsx.AssertExistsAll(filepath.Join(root, "package.json")) != nil {
		if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}"), 0777); err != nil {
			return nil, errors.Wrap(err, "creating default package.json")
		}
	}

	isYarn := fsx.AssertExistsAll(filepath.Join(root, "yarn.lock")) == nil
	var cmd *exec.Cmd
	if isYarn {
		cmd = exec.CommandContext(ctx, "yarn", "add", "-D", "@types/node")
	} else {
		cmd = exec.CommandContext(ctx, "npm", "install", "--save-dev", "@types/node")
	}
	cmd.Dir = workdir
	if err := cmd.Run(); err != nil {
		return nil, errors.New("failed to add @types/node dependency")
	}

	cmd = exec.CommandContext(ctx, "tsc", build.NodeTscArgs(".", opts.KindOptions)...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log(strings.TrimSpace(string(out)))
		logger.Debug("\nCommand: %s", strings.Join(cmd.Args, " "))

		return nil, errors.Errorf("failed to compile %s", opts.Path)
	}

	pv, err := json.Marshal(opts.ParamValues)
	if err != nil {
		return nil, errors.Wrap(err, "serializing param values")
	}

	return []string{"node", filepath.Join(root, ".airplane/dist/.airplane/shim.js"), string(pv)}, nil
}

// checkTscInstalled will error if the tsc CLI is not installed.
//
// TODO: consider either a) auto-installing tsc or b) packaging it
// with the airplane CLI. The latter would be ideal, since we could
// enforce that the correct version of tsc is used.
func checkTscInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "tsc", "--version")
	if err := cmd.Run(); err != nil {
		return errors.New(heredoc.Doc(`
			It looks like the typescript CLI (tsc) is not installed.

			You can install it with:
			  npm install -g typescript
			  tsc --version
			
			See also: https://www.typescriptlang.org/download
		`))
	}

	return nil
}

// checkNodeVersion compares the major version of the currently installed
// node binary with that of the configured task and logs a warning if they
// do not match.
func checkNodeVersion(ctx context.Context, opts api.KindOptions) {
	nodeVersion, ok := opts["nodeVersion"].(string)
	if !ok {
		return
	}

	v, err := semver.ParseTolerant(nodeVersion)
	if err != nil {
		logger.Debug("Unable to parse node version (%s): ignoring", nodeVersion)
		return
	}

	cmd := exec.CommandContext(ctx, "node", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("failed to check node version: is node installed?")
		return
	}

	if !strings.HasPrefix(string(out), fmt.Sprintf("v%d", v.Major)) {
		logger.Warning("Your local version of Node (%s) does not match the version your task is configured to run against (v%s).", strings.TrimSpace(string(out)), v)
	}
}
