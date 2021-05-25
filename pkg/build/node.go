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
		HasPackageJSON bool
		HasPackageLock bool
		HasYarnLock    bool
		Shim           string
		IsTS           bool
	}{
		HasPackageJSON: exist(filepath.Join(root, "package.json")) == nil,
		HasPackageLock: exist(filepath.Join(root, "package-lock.json")) == nil,
		HasYarnLock:    exist(filepath.Join(root, "yarn.lock")) == nil,
		IsTS:           strings.HasSuffix(entrypoint, ".ts"),
	}

	cfg.Base, err = getBaseNodeImage(args["nodeVersion"])
	if err != nil {
		return "", err
	}

	relimport, err := filepath.Rel(root, entrypoint)
	if err != nil {
		return "", errors.Wrap(err, "entrypoint is not inside of root")
	}
	// Remove the `.ts` suffix if one exists, since tsc doesn't accept
	// import paths with `.ts` endings. `.js` endings are fine.
	relimport = strings.TrimSuffix(relimport, ".ts")

	shim := `// This file includes a shim that will execute your task code.
import task from "../` + relimport + `"

async function main() {
	if (process.argv.length !== 3) {
		console.log("airplane_output:error " + JSON.stringify({ "error": "Expected to receive a single argument (via {{JSON}}). Task CLI arguments may be misconfigured." }))
		process.exit(1)
	}
	
	try {
		await task(JSON.parse(process.argv[2]))
	} catch (err) {
		console.error(err)
		console.log("airplane_output:error " + JSON.stringify({ "error": String(err) }))
		process.exit(1)
	}
}

main()`
	// To inline the shim into a Dockerfile, insert `\n\` characters:
	cfg.Shim = strings.Join(strings.Split(shim, "\n"), "\\n\\\n")

	// TODO: do we want to support buildDir and buildCommand still?
	return templatize(`
		FROM {{.Base}}

		WORKDIR /airplane

		# Support setting BUILD_NPM_RC or BUILD_NPM_TOKEN to configure private registry auth
		ARG BUILD_NPM_RC
		ARG BUILD_NPM_TOKEN
		RUN [ -z "${BUILD_NPM_RC}" ] || echo "${BUILD_NPM_RC}" > .npmrc
		RUN [ -z "${BUILD_NPM_TOKEN}" ] || echo "//registry.npmjs.org/:_authToken=${BUILD_NPM_TOKEN}" > .npmrc

		RUN npm install -g typescript@4.2

		{{if .HasPackageJSON}}
		COPY package.json .
		{{else}}
		RUN echo '{}' > package.json
		{{end}}

		{{if .HasPackageLock}}
		RUN npm install package-lock.json
		{{else if .HasYarnLock}}
		RUN yarn install
		{{end}}

		COPY . .

		RUN mkdir -p .airplane-build/dist && \
			echo '{{.Shim}}' > .airplane-build/shim.{{if .IsTS}}ts{{else}}js{{end}} && \
			cp package.json .airplane-build/dist/package.json && \
			tsc \
				--allowJs \
				--module commonjs \
				--target es2020 \
				--lib es2020 \
				--esModuleInterop \
				--outDir .airplane-build/dist \
				--rootDir . \
				.airplane-build/shim.{{if .IsTS}}ts{{else}}js{{end}}
		ENTRYPOINT ["node", ".airplane-build/dist/.airplane-build/shim.js"]
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
