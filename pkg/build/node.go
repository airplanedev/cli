package build

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
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
		cmds = append(cmds, `npm install -g typescript@4.1`)
		cmds = append(cmds, `[-f tsconfig.json] || echo '{"include": ["*", "**/*"], "exclude": ["node_modules"]}' >tsconfig.json`)
		cmds = append(cmds, `rm -rf .airplane-build/ && tsc --outDir .airplane-build --rootDir .`)
		entrypoint = "/airplane/.airplane-build/" + strings.TrimSuffix(entrypoint, ".ts") + ".js"

	case "javascript":
		entrypoint = "/airplane/" + entrypoint

	default:
		return "", fmt.Errorf("build: unknown language %q, it must be javascript or tyescript", lang)

	}

	// Dockerfile template.
	t, err := template.New("node").Parse(`
FROM {{ .Base }}

WORKDIR /airplane
COPY . /airplane
{{ range .Commands }}
RUN {{ . }}
{{ end }}

ENTRYPOINT ["node", "{{ .Main }}"]
`)
	if err != nil {
		return "", err
	}

	var data struct {
		Base     string
		Commands []string
		Main     string
	}
	data.Base = expandNodeVersion(version)
	data.Commands = cmds
	data.Main = entrypoint

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// expandNodeVersion returns a pinned minor version of Node to use
func expandNodeVersion(version string) string {
	switch version {
	case "":
		// If empty, use default of 15
		fallthrough
	case "15":
		// 15.14-buster
		return "node@sha256:0227179cbab2998464b8c9ec77812996aec73b7a4cceb0fcacc49bfff8b8cf8d"
	case "14":
		return "node:14.16-buster"
	case "12":
		return "node:12.22-buster"
	default:
		// Assume the version is already a more-specific version - default to just returning it back
		return "node:" + version + "-buster"
	}
}
