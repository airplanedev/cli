package build

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// node creates a dockerfile for Node (typescript/javascript).
func node(root string, args Args) (string, error) {
	var err error

	// For backwards compatibility, continue to build old Node tasks
	// in the same way. Tasks built with the latest CLI will set
	// shim=true which enables the new code path.
	if shim := args["shim"]; shim != "true" {
		return nodeLegacyBuilder(root, args)
	}

	// Assert that the entrypoint file exists:
	entrypoint := filepath.Join(root, args["entrypoint"])
	if err := exist(entrypoint); err != nil {
		return "", err
	}

	cfg := struct {
		Workdir        string
		Base           string
		HasPackageJSON bool
		HasPackageLock bool
		HasYarnLock    bool
		Shim           string
		IsTS           bool
		TscTarget      string
		TscLib         string
	}{
		Workdir:        args["workdir"],
		HasPackageJSON: exist(filepath.Join(root, "package.json")) == nil,
		HasPackageLock: exist(filepath.Join(root, "package-lock.json")) == nil,
		HasYarnLock:    exist(filepath.Join(root, "yarn.lock")) == nil,
		// https://github.com/tsconfig/bases/blob/master/bases/node16.json
		TscTarget: "es2020",
		TscLib:    "es2020",
	}

	if !strings.HasPrefix(cfg.Workdir, "/") {
		cfg.Workdir = "/" + cfg.Workdir
	}

	nodeVersion := args["nodeVersion"]
	// For node version 12, 12.x, etc., we need to change the ECMAScript target.
	// https://github.com/tsconfig/bases/blob/master/bases/node12.json
	if strings.HasPrefix(nodeVersion, "12") {
		cfg.TscTarget = "es2019"
		cfg.TscLib = "es2019"
	}
	cfg.Base, err = getBaseNodeImage(nodeVersion)
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

	// The following Dockerfile can build both JS and TS tasks. In general, we're
	// aiming for recent EC202x support and for support for import/export syntax.
	// The former is easier, since recent versions of Node have excellent coverage
	// of the ECMAScript spec. The latter could be achieved through ECMAScript
	// modules (ESM), but those are not well-supported within the Node community.
	// Basic functionality of ESM is also still in the experimental stage, such as
	// module resolution for relative paths (f.e. ./main.js vs. ./main). Therefore,
	// we have to fallback to a separate build step to offer import/export support.
	// We have a few options -- f.e. babel or esbuild -- but the easiest is simply
	// using the tsc compiler for JS projects, too.
	//
	// Down the road, we may want to give customers more control over this build process
	// in which case we could introduce an extra step for performing build commands.
	return templatize(`
		FROM {{.Base}}

		WORKDIR /airplane{{.Workdir}}

		# Support setting BUILD_NPM_RC or BUILD_NPM_TOKEN to configure private registry auth
		ARG BUILD_NPM_RC
		ARG BUILD_NPM_TOKEN
		RUN [ -z "${BUILD_NPM_RC}" ] || echo "${BUILD_NPM_RC}" > .npmrc
		RUN [ -z "${BUILD_NPM_TOKEN}" ] || echo "//registry.npmjs.org/:_authToken=${BUILD_NPM_TOKEN}" > .npmrc

		RUN npm install -g typescript@4.2

		COPY . /airplane

		{{if not .HasPackageJSON}}
		RUN echo '{}' > /airplane/package.json
		{{end}}

		{{if .HasPackageLock}}
		RUN npm install package-lock.json && npm install --save-dev @types/node
		{{else if .HasYarnLock}}
		RUN yarn --frozen-lockfile --non-interactive && yarn add -D @types/node
		{{else}}
		RUN npm install --save-dev @types/node
		{{end}}

		RUN mkdir -p /airplane/.airplane-build/dist && \
			echo '{{.Shim}}' > /airplane/.airplane-build/shim.ts && \
			tsc \
				--allowJs \
				--module commonjs \
				--target {{.TscTarget}} \
				--lib {{.TscLib}} \
				--esModuleInterop \
				--outDir /airplane/.airplane-build/dist \
				--rootDir /airplane \
				--skipLibCheck \
				/airplane/.airplane-build/shim.ts
		ENTRYPOINT ["node", "/airplane/.airplane-build/dist/.airplane-build/shim.js"]
	`, cfg)
}

// nodeLegacyBuilder creates a dockerfile for Node (typescript/javascript).
//
// TODO(amir): possibly just run `npm start` instead of exposing lots
// of options to users?
func nodeLegacyBuilder(root string, args Args) (string, error) {
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
		version = "16"
	}
	v, err := GetVersion(NameNode, version)
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
