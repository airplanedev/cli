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

	// Default version.
	if version == "" {
		version = "15.8"
	}

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
FROM node:{{ .NodeVersion }}-buster

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
		NodeVersion string
		Commands    []string
		Main        string
	}
	data.NodeVersion = version
	data.Commands = cmds
	data.Main = entrypoint

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
