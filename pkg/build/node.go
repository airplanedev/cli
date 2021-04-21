package build

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// Node creates a dockerfile for Node (typescript/javascript).
//
// TODO(amir): possibly just run `npm start` instead of exposing lots
// of options to users?
func node(root string, args Args) (string, error) {
	var entrypoint = args["entrypoint"]
	var main = filepath.Join(root, entrypoint)
	var deps = filepath.Join(root, "package.json")
	var yarnlock = filepath.Join(root, "yarn.lock")
	var pkglock = filepath.Join(root, "package-lock.json")
	var version = args["nodeVersion"]
	var lang = args["language"]
	// `workdir` is fixed usually - `buildWorkdir` is a subdirectory of `workdir` if there's
	// `buildCommand` and is ultimately where `entrypoint` is run from.
	var buildCommand = args["buildCommand"]
	var buildDir = args["buildDir"]
	var workdir = "/airplane"
	var buildWorkdir = "/airplane"
	var cmds []string

	// Make sure that entrypoint and `package.json` exist.
	if err := exist(main, deps); err != nil {
		return "", err
	}

	// Determine the install command to use.
	if err := exist(pkglock); err == nil {
		cmds = append(cmds, `npm install package-lock.json`)
	} else if err := exist(yarnlock); err == nil {
		cmds = append(cmds, `yarn install`)
	}

	// Language specific.
	switch lang {
	case "typescript":
		if buildDir == "" {
			buildDir = ".airplane-build"
		}
		cmds = append(cmds, `npm install -g typescript@4.1`)
		cmds = append(cmds, `[ -f tsconfig.json ] || echo '{"include": ["*", "**/*"], "exclude": ["node_modules"]}' >tsconfig.json`)
		cmds = append(cmds, fmt.Sprintf(`rm -rf %s && tsc --outDir %s --rootDir .`, buildDir, buildDir))
		if buildCommand != "" {
			// It's not totally expected, but if you do set buildCommand we'll run it after tsc
			cmds = append(cmds, buildCommand)
		}
		buildWorkdir = path.Join(workdir, buildDir)
		// If entrypoint ends in .ts, replace it with .js
		entrypoint = strings.TrimSuffix(entrypoint, ".ts") + ".js"
	case "javascript":
		if buildCommand != "" {
			cmds = append(cmds, buildCommand)
		}
		if buildDir != "" {
			buildWorkdir = path.Join(workdir, buildDir)
		}
	default:
		return "", errors.Errorf("build: unknown language %q, expected \"javascript\" or \"typescript\"", lang)
	}
	entrypoint = path.Join(buildWorkdir, entrypoint)

	// Dockerfile template.
	t, err := template.New("node").Parse(`
FROM {{ .Base }}

WORKDIR {{ .Workdir }}
COPY . {{ .Workdir }}
{{ range .Commands }}
RUN {{ . }}
{{ end }}

WORKDIR {{ .BuildWorkdir }}
ENTRYPOINT ["node", "{{ .Main }}"]
`)
	if err != nil {
		return "", err
	}

	if version == "" {
		version = "15"
	}
	v, err := GetVersion(BuilderNameNode, version)
	if err != nil {
		return "", err
	}
	base := v.String()
	if base == "" {
		// Assume the version is already a more-specific version - default to just returning it back
		base = "node:" + version + "-buster"
	}

	var buf strings.Builder
	if err := t.Execute(&buf, struct {
		Base         string
		Workdir      string
		BuildWorkdir string
		Commands     []string
		Main         string
	}{
		Base:         base,
		Workdir:      workdir,
		BuildWorkdir: buildWorkdir,
		Commands:     cmds,
		Main:         entrypoint,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
