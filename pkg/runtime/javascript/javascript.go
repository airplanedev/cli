package javascript

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/utils"
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

// Workdir picks the working directory for commands to be executed from.
//
// For JS, that is the nearest parent directory containing a `package.json`.
func (r Runtime) Workdir(path string) (string, error) {
	if p, ok := fsx.Find(path, "package.json"); ok {
		return p, nil
	}

	return "", errors.New("a package.json could not be found")
}

// Root picks which directory to use as the root of a task's code.
// All code in that directory will be available at runtime.
//
// For JS, this is usually just the workdir. However, this can be overridden
// with the `airplane.root` property in the `package.json`.
func (r Runtime) Root(path string) (string, error) {
	// By default, the root is the workdir.
	root, err := r.Workdir(path)
	if err != nil {
		return "", err
	}

	// Unless the root is overridden with an `airplane.root` field
	// in a `package.json`.
	pkgjson := filepath.Join(root, "package.json")
	buf, err := os.ReadFile(pkgjson)
	if err != nil {
		return "", errors.Wrapf(err, "javascript: reading %s", pkgjson)
	}

	var pkg struct {
		Settings runtime.Settings `json:"airplane"`
	}

	if err := json.Unmarshal(buf, &pkg); err != nil {
		return "", fmt.Errorf("javascript: reading %s - %w", root, err)
	}

	if pkgjsonRoot := pkg.Settings.Root; pkgjsonRoot != "" {
		return filepath.Join(root, pkgjsonRoot), nil
	}

	return root, nil
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

	if err := os.Mkdir(filepath.Join(root, ".airplane"), os.ModeDir|0777); err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "creating .airplane directory")
	}

	entrypoint, err := filepath.Rel(root, opts.Path)
	if err != nil {
		return nil, errors.Wrap(err, "entrypoint is not within the task root")
	}
	shim, err := build.NodeShim(entrypoint)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filepath.Join(root, ".airplane/shim.ts"), []byte(shim), 0644); err != nil {
		return nil, errors.Wrap(err, "writing shim file")
	}

	if err := os.RemoveAll(filepath.Join(root, ".airplane/dist")); err != nil {
		return nil, errors.Wrap(err, "cleaning dist folder")
	}

	// Confirm we have a `package.json`, otherwise we might install shim dependencies
	// in the wrong folder.
	hasPkgJSON := fsx.AssertExistsAll(filepath.Join(root, "package.json")) == nil
	if !hasPkgJSON {
		return nil, errors.New("a package.json is missing")
	}

	if !build.HasNodeShimDeps(root) {
		if err := installShimDeps(ctx, root, opts.Path); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, "npx", append([]string{"-p", "typescript", "--no", "tsc", "--"}, build.NodeTscArgs(".", opts.KindOptions)...)...)
	cmd.Dir = root
	logger.Debug("Running %s (in %s)", logger.Bold(strings.Join(cmd.Args, " ")), root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log(strings.TrimSpace(string(out)))

		return nil, errors.Errorf("failed to compile %s", opts.Path)
	}
	logger.Debug("Compiled JS in %s", logger.Bold(time.Since(start).String()))

	pv, err := json.Marshal(opts.ParamValues)
	if err != nil {
		return nil, errors.Wrap(err, "serializing param values")
	}

	return []string{"node", filepath.Join(root, ".airplane/dist/.airplane/shim.js"), string(pv)}, nil
}

func installShimDeps(ctx context.Context, root, path string) error {
	isYarn := fsx.AssertExistsAll(filepath.Join(root, "yarn.lock")) == nil
	var cmd *exec.Cmd
	if isYarn {
		cmd = exec.CommandContext(ctx, "yarn", "add", "-D", "@types/node")
	} else {
		cmd = exec.CommandContext(ctx, "npm", "install", "--save-dev", "@types/node")
	}
	cmd.Dir = filepath.Dir(path)

	// Confirm with the user before installing the shim dependencies.
	if utils.CanPrompt() {
		logger.Log("Airplane needs to run %s before it can build your task.", logger.Bold(strings.Join(cmd.Args, " ")))
		confirmed, err := utils.Confirm("Run now?")
		if err != nil {
			return err
		}
		if !confirmed {
			return errors.New("unable to run without shim dependencies")
		}
	}

	logger.Debug("Running %s", logger.Bold(strings.Join(cmd.Args, " ")))
	if err := cmd.Run(); err != nil {
		return errors.New("failed to add shim dependencies")
	}

	return nil
}

// checkTscInstalled will verify that the Typescript CLI is installed
// and confirm with the user if they are okay with us auto-installing it.
func checkTscInstalled(ctx context.Context) error {
	// note: --no will prevent installing typescript if not already installed.
	cmd := exec.CommandContext(ctx, "npx", "-p", "typescript", "--no", "tsc", "--", "--version")
	logger.Debug("Running %s", logger.Bold(strings.Join(cmd.Args, " ")))
	if out, err := cmd.CombinedOutput(); err == nil {
		logger.Debug("TypeScript version: %s", strings.TrimPrefix(strings.TrimSpace(string(out)), "Version "))
		// tsc is installed, return early
		return nil
	}

	// Typescript is not installed. Confirm with the user if they are
	// okay with installing it.
	cmd = exec.CommandContext(ctx, "npx", "-p", "typescript", "--yes", "tsc", "--version")
	if utils.CanPrompt() {
		logger.Log("Airplane needs to run %s to install the TypeScript CLI.", logger.Bold(strings.Join(cmd.Args, " ")))
		confirmed, err := utils.Confirm("Run now?")
		if err != nil {
			return err
		}
		if !confirmed {
			return errors.New("unable to run without the TypeScript CLI")
		}
	}

	logger.Debug("Running %s", logger.Bold(strings.Join(cmd.Args, " ")))
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "installing tsc")
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
	logger.Debug("Running %s", logger.Bold(strings.Join(cmd.Args, " ")))
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("failed to check node version: is node installed?")
		return
	}

	logger.Debug("node version: %s", strings.TrimSpace(string(out)))
	if !strings.HasPrefix(string(out), fmt.Sprintf("v%d", v.Major)) {
		logger.Warning("Your local version of Node (%s) does not match the version your task is configured to run against (v%s).", strings.TrimSpace(string(out)), v)
	}

	cmd = exec.CommandContext(ctx, "npx", "--version")
	logger.Debug("Running %s", logger.Bold(strings.Join(cmd.Args, " ")))
	out, err = cmd.CombinedOutput()
	if err != nil {
		logger.Debug("failed to check npx version: are you running a recent enough version of node?")
		return
	}

	logger.Debug("npx version: %s", strings.TrimSpace(string(out)))
}
