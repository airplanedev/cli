package build

import (
	_ "embed"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

// node creates a dockerfile for Node (typescript/javascript).
func node(root string, options api.KindOptions) (string, error) {
	var err error

	// For backwards compatibility, continue to build old Node tasks
	// in the same way. Tasks built with the latest CLI will set
	// shim=true which enables the new code path.
	if shim, ok := options["shim"].(string); !ok || shim != "true" {
		return nodeLegacyBuilder(root, options)
	}

	// Assert that the entrypoint file exists:
	entrypoint, _ := options["entrypoint"].(string)
	entrypoint = filepath.Join(root, entrypoint)
	if err := fsx.AssertExistsAll(entrypoint); err != nil {
		return "", err
	}

	workdir, _ := options["workdir"].(string)
	cfg := struct {
		Workdir        string
		Base           string
		HasPackageJSON bool
		HasPackageLock bool
		HasYarnLock    bool
		Shim           string
		IsTS           bool
		TscArgs        string
	}{
		Workdir:        workdir,
		HasPackageJSON: fsx.AssertExistsAll(filepath.Join(root, "package.json")) == nil,
		HasPackageLock: fsx.AssertExistsAll(filepath.Join(root, "package-lock.json")) == nil,
		HasYarnLock:    fsx.AssertExistsAll(filepath.Join(root, "yarn.lock")) == nil,
		TscArgs:        strings.Join(NodeTscArgs("/airplane", options), " \\\n"),
	}

	if !strings.HasPrefix(cfg.Workdir, "/") {
		cfg.Workdir = "/" + cfg.Workdir
	}

	nodeVersion, _ := options["nodeVersion"].(string)
	cfg.Base, err = getBaseNodeImage(nodeVersion)
	if err != nil {
		return "", err
	}

	shim, err := NodeShim(root, entrypoint)
	if err != nil {
		return "", err
	}
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
	return applyTemplate(`
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

		{{if .HasYarnLock}}
		RUN yarn --frozen-lockfile --non-interactive && yarn add -D @types/node
		{{else if .HasPackageLock}}
		RUN npm install && npm install --save-dev @types/node
		{{else}}
		RUN npm install --save-dev @types/node
		{{end}}

		RUN mkdir -p /airplane/.airplane/dist && \
			echo '{{.Shim}}' > /airplane/.airplane/shim.ts && \
			tsc {{.TscArgs}}
		ENTRYPOINT ["node", "/airplane/.airplane/dist/.airplane/shim.js"]
	`, cfg)
}

//go:embed node-shim.ts
var nodeShim string

func NodeShim(root, entrypoint string) (string, error) {
	importPath, err := filepath.Rel(root, entrypoint)
	if err != nil {
		return "", errors.Wrap(err, "entrypoint is not inside of root")
	}
	// Remove the `.ts` suffix if one exists, since tsc doesn't accept
	// import paths with `.ts` endings. `.js` endings are fine.
	importPath = strings.TrimSuffix(importPath, ".ts")

	shim, err := applyTemplate(nodeShim, struct {
		ImportPath string
	}{
		ImportPath: importPath,
	})
	if err != nil {
		return "", errors.Wrap(err, "templating shim")
	}

	return shim, nil
}

func NodeTscArgs(root string, opts api.KindOptions) []string {
	// https://github.com/tsconfig/bases/blob/master/bases/node16.json
	tscTarget := "es2020"
	tscLib := "es2020"
	if strings.HasPrefix(opts["nodeVersion"].(string), "12") {
		tscTarget = "es2019"
		tscLib = "es2019"
	}

	return []string{
		"--allowJs",
		"--module", "commonjs",
		"--target", tscTarget,
		"--lib", tscLib,
		"--esModuleInterop",
		"--outDir", filepath.Join(root, ".airplane/dist"),
		"--rootDir", root,
		"--skipLibCheck",
		"--pretty",
		filepath.Join(root, ".airplane/shim.ts"),
	}
}

// nodeLegacyBuilder creates a dockerfile for Node (typescript/javascript).
//
// TODO(amir): possibly just run `npm start` instead of exposing lots
// of options to users?
func nodeLegacyBuilder(root string, options api.KindOptions) (string, error) {
	entrypoint, _ := options["entrypoint"].(string)
	main := filepath.Join(root, entrypoint)
	deps := filepath.Join(root, "package.json")
	yarnlock := filepath.Join(root, "yarn.lock")
	pkglock := filepath.Join(root, "package-lock.json")
	lang, _ := options["language"].(string)
	// `workdir` is fixed usually - `buildWorkdir` is a subdirectory of `workdir` if there's
	// `buildCommand` and is ultimately where `entrypoint` is run from.
	buildCommand, _ := options["buildCommand"].(string)
	buildDir, _ := options["buildDir"].(string)
	workdir := "/airplane"
	buildWorkdir := "/airplane"
	cmds := []string{}
	nodeVersion, _ := options["nodeVersion"].(string)

	// Make sure that entrypoint and `package.json` exist.
	if err := fsx.AssertExistsAll(main, deps); err != nil {
		return "", err
	}

	// Determine the install command to use.
	if err := fsx.AssertExistsAll(pkglock); err == nil {
		cmds = append(cmds, `npm install package-lock.json`)
	} else if err := fsx.AssertExistsAll(yarnlock); err == nil {
		cmds = append(cmds, `yarn install`)
	}

	// Language specific.
	switch lang {
	case "typescript":
		if buildDir == "" {
			buildDir = ".airplane"
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

	baseImage, err := getBaseNodeImage(nodeVersion)
	if err != nil {
		return "", err
	}

	return applyTemplate(`
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
