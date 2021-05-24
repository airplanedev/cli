package build

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// node creates a dockerfile for Node (typescript/javascript).
func node(root string, args Args) (string, error) {
	var err error

	// For backwards compatibility, continue to build old Node tasks
	// in the same way. Tasks built with the latest CLI will set
	// shim=true which enables the new code path.
	if shim := args["shim"]; shim != "true" {
		return nodeOld(root, args)
	}

	// Assert that the entrypoint file exists:
	entrypoint := filepath.Join(root, args["entrypoint"])
	if err := exist(entrypoint); err != nil {
		return "", err
	}

	cfg := struct {
		Base           string
		Entrypoint     string
		HasPackageJSON bool
		HasPackageLock bool
		HasYarnLock    bool
		CreateTSConfig bool
		HasTSConfig    bool
		RunTSC         bool
	}{
		HasPackageJSON: exist(filepath.Join(root, "package.json")) == nil,
		HasPackageLock: exist(filepath.Join(root, "package-lock.json")) == nil,
		HasYarnLock:    exist(filepath.Join(root, "yarn.lock")) == nil,
	}

	cfg.Base, err = getBaseNodeImage(args["nodeVersion"])
	if err != nil {
		return "", err
	}

	if args["language"] == "typescript" {
		cfg.RunTSC = true
		cfg.HasTSConfig = exist(filepath.Join(root, "tsconfig.json")) == nil
		// If a tsconfig.json was not provided, insert a default one:
		cfg.CreateTSConfig = !cfg.HasTSConfig
		// Point the entrypoint at the compiled JS version of the entrypoint.
		// TODO: consider using ts-node
		cfg.Entrypoint = filepath.Join(".airplane-build", strings.TrimSuffix(entrypoint, ".ts")+".js")
	} else {
		cfg.Entrypoint = entrypoint
	}

	// TODO: do we want to support buildDir and buildCommand still?
	return templatize(`
		FROM {{.Base}}

		WORKDIR /airplane

		# Support setting BUILD_NPM_RC or BUILD_NPM_TOKEN to configure private registry auth
		ARG BUILD_NPM_RC
		ARG BUILD_NPM_TOKEN
		RUN [ -z "${BUILD_NPM_RC}" ] || echo "${BUILD_NPM_RC}" > .npmrc
		RUN [ -z "${BUILD_NPM_TOKEN}" ] || echo "//registry.npmjs.org/:_authToken=${BUILD_NPM_TOKEN}" > .npmrc

		{{if .RunTSC}}
		RUN npm install -g typescript@4.2
		{{end}}

		{{if .HasPackageJSON}}
		COPY package.json .
		{{else}}
		RUN echo '{"type":"module"}' > package.json
		{{end}}

		{{if .HasTSConfig}}
		COPY tsconfig.json .
		{{else if .CreateTSConfig}}
		RUN echo '{"include": ["*", "**/*"], "exclude": ["node_modules"]}' > tsconfig.json
		{{end}}

		{{if .HasPackageLock}}
		RUN npm install package-lock.json
		{{else if .HasYarnLock}}
		RUN yarn install
		{{end}}

		COPY . .

		{{if .RunTSC}}
		RUN mkdir .airplane-build && tsc --outDir .airplane-build --rootDir .
		{{end}}

		ENTRYPOINT ["node", "--input-type=module", "{{ .Main }}"]
	`, cfg)
}

func templatize(t string, data interface{}) (string, error) {
	tmpl, err := template.New("airplane").Parse(t)
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// nodeOld creates a dockerfile for Node (typescript/javascript).
//
// TODO(amir): possibly just run `npm start` instead of exposing lots
// of options to users?
func nodeOld(root string, args Args) (string, error) {
	var entrypoint = args["entrypoint"]
	var main = filepath.Join(root, entrypoint)
	var deps = filepath.Join(root, "package.json")
	var yarnlock = filepath.Join(root, "yarn.lock")
	var pkglock = filepath.Join(root, "package-lock.json")
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

	baseImage, err := getBaseNodeImage(args["nodeVersion"])
	if err != nil {
		return "", err
	}

	return templatize(`
		FROM {{ .Base }}
		
		WORKDIR {{ .Workdir }}
		
		# Support setting BUILD_NPM_RC or BUILD_NPM_TOKEN to configure private registry auth
		ARG BUILD_NPM_RC
		ARG BUILD_NPM_TOKEN
		RUN [ -z "${BUILD_NPM_RC}" ] || echo "${BUILD_NPM_RC}" > .npmrc
		RUN [ -z "${BUILD_NPM_TOKEN}" ] || echo "//registry.npmjs.org/:_authToken=${BUILD_NPM_TOKEN}" > .npmrc
		
		COPY . {{ .Workdir }}
		{{ range .Commands }}
		RUN {{ . }}
		{{ end }}
		
		WORKDIR {{ .BuildWorkdir }}
		ENTRYPOINT ["node", "{{ .Main }}"]
	`, struct {
		Base         string
		Workdir      string
		BuildWorkdir string
		Commands     []string
		Main         string
	}{
		Base:         baseImage,
		Workdir:      workdir,
		BuildWorkdir: buildWorkdir,
		Commands:     cmds,
		Main:         entrypoint,
	})
}

func getBaseNodeImage(version string) (string, error) {
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

	return base, nil
}
