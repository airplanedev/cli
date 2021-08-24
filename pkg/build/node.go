package build

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils/pointers"
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
	if err := fsx.AssertExistsAll(filepath.Join(root, entrypoint)); err != nil {
		return "", err
	}

	workdir, _ := options["workdir"].(string)
	cfg := struct {
		Workdir               string
		Base                  string
		HasPackageJSON        bool
		IsYarn                bool
		InlineShim            string
		InlineTSConfig        string
		InlineShimPackageJSON string
	}{
		Workdir:        workdir,
		HasPackageJSON: fsx.AssertExistsAll(filepath.Join(root, "package.json")) == nil,
		IsYarn:         fsx.AssertExistsAll(filepath.Join(root, "yarn.lock")) == nil,
	}

	if !strings.HasPrefix(cfg.Workdir, "/") {
		cfg.Workdir = "/" + cfg.Workdir
	}

	nodeVersion, _ := options["nodeVersion"].(string)
	cfg.Base, err = getBaseNodeImage(nodeVersion)
	if err != nil {
		return "", err
	}

	pjson, err := GenShimPackageJSON()
	if err != nil {
		return "", err
	}
	cfg.InlineShimPackageJSON = inlineString(string(pjson))

	shim, err := NodeShim(entrypoint)
	if err != nil {
		return "", err
	}
	cfg.InlineShim = inlineString(shim)

	tsconfig, err := GenTSConfig(root, filepath.Join(root, entrypoint), options)
	if err != nil {
		return "", err
	}
	cfg.InlineTSConfig = inlineString(string(tsconfig))

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
	return applyTemplate(heredoc.Doc(`
		FROM {{.Base}}

		WORKDIR /airplane{{.Workdir}}

		# Support setting BUILD_NPM_RC or BUILD_NPM_TOKEN to configure private registry auth
		ARG BUILD_NPM_RC
		ARG BUILD_NPM_TOKEN
		RUN [ -z "${BUILD_NPM_RC}" ] || echo "${BUILD_NPM_RC}" > .npmrc
		RUN [ -z "${BUILD_NPM_TOKEN}" ] || echo "//registry.npmjs.org/:_authToken=${BUILD_NPM_TOKEN}" > .npmrc

		RUN npm install -g typescript@4.2
		COPY . /airplane

		RUN mkdir -p /airplane/.airplane && \
			cd /airplane/.airplane && \
			{{.InlineShimPackageJSON}} > package.json && \
			npm install

		{{if not .HasPackageJSON}}
		RUN echo '{}' > /airplane/package.json
		{{end}}

		{{if .IsYarn}}
		RUN yarn --non-interactive
		{{else}}
		RUN npm install
		{{end}}

		RUN {{.InlineShim}} > /airplane/.airplane/shim.ts && \
			{{.InlineTSConfig}} > /airplane/.airplane/tsconfig.json && \
			tsc --pretty -p /airplane/.airplane
		ENTRYPOINT ["node", "/airplane/.airplane/dist/.airplane/shim.js"]
	`), cfg)
}

func GenShimPackageJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Dependencies map[string]string `json:"dependencies"`
	}{
		Dependencies: map[string]string{
			"@types/node": "^16",
		},
	})
	return b, errors.Wrap(err, "generating shim dependencies")
}

// GenTSConfig generates a `tsconfig.json` that can be placed in `<root>/.airplane/tsconfig.json`.
//
// If a user-provided tsconfig.json is found, in between the root and entrypoint directories,
// then the generated `tsconfig.json` will instruct tsc to read that.
func GenTSConfig(root string, entrypoint string, opts api.KindOptions) ([]byte, error) {
	// https://www.typescriptlang.org/tsconfig
	type CompilerOptions struct {
		Target          string                     `json:"target,omitempty"`
		Lib             []string                   `json:"lib,omitempty"`
		AllowJS         *bool                      `json:"allowJs,omitempty"`
		Module          string                     `json:"module,omitempty"`
		ESModuleInterop *bool                      `json:"esModuleInterop,omitempty"`
		OutDir          string                     `json:"outDir"`
		RootDir         string                     `json:"rootDir"`
		SkipLibCheck    *bool                      `json:"skipLibCheck,omitempty"`
		Paths           map[string]json.RawMessage `json:"paths,omitempty"`
	}
	type TSConfig struct {
		CompilerOptions CompilerOptions `json:"compilerOptions"`
		Files           []string        `json:"files"`
		Extends         string          `json:"extends,omitempty"`
	}

	tsconfig := TSConfig{
		// The following configuration takes precedence over a user-provided tsconfig.
		// All other tsconfig fields should be set to `omitempty` so that they can be
		// overridden by a user-provided tsconfig.
		CompilerOptions: CompilerOptions{
			// This tsconfig is placed in `<root>/.airplane/tsconfig.json`
			RootDir: "..",
			// Placed compiled files into `<root>/.airplane/dist`
			OutDir: "./dist",
		},
		// `shim.ts` is our entrypoint. When we point tsc at this tsconfig, it will
		// compile shim.ts and all of its imported files.
		Files: []string{"./shim.ts"},
	}

	// Check if the user provided their own tsconfig. Use the tsconfig closest to the user's entrypoint.
	var utsc TSConfig
	if p, ok := fsx.FindUntil(filepath.Dir(entrypoint), root, "tsconfig.json"); ok {
		p = filepath.Join(p, "tsconfig.json")
		// Read the contents of the user's tsconfig and warn about any unsupported behavior:
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, errors.Wrap(err, "reading user-provided tsconfig")
		}
		logger.Debug("Found tsconfig.json at %s: %+v", p, strings.TrimSpace(string(content)))
		if err := json.Unmarshal(content, &utsc); err != nil {
			return nil, errors.Wrap(err, "invalid tsconfig.json")
		}
		if len(utsc.CompilerOptions.Paths) > 0 {
			logger.Warning("Detected a tsconfig.json with path aliases which are not supported on Airplane yet.")
		}

		rp, err := filepath.Rel(filepath.Join(root, ".airplane"), p)
		if err != nil {
			return nil, errors.Wrap(err, "creating relative tsconfig path")
		}
		tsconfig.Extends = rp
	}

	// Apply defaults to a few of the tsconfig fields, but let the user override
	// them from their tsconfig.json:
	if utsc.CompilerOptions.AllowJS == nil {
		tsconfig.CompilerOptions.AllowJS = pointers.Bool(true)
	}
	if utsc.CompilerOptions.Module == "" {
		tsconfig.CompilerOptions.Module = "commonjs"
	}
	if utsc.CompilerOptions.ESModuleInterop == nil {
		tsconfig.CompilerOptions.ESModuleInterop = pointers.Bool(true)
	}
	if utsc.CompilerOptions.SkipLibCheck == nil {
		tsconfig.CompilerOptions.SkipLibCheck = pointers.Bool(true)
	}

	target := "es2020"
	if opts != nil && strings.HasPrefix(opts["nodeVersion"].(string), "12") {
		// For Node 12 (the earliest version of Node we support), we need to compile to an
		// older version of ECMAScript.
		target = "es2019"
	}
	if utsc.CompilerOptions.Target == "" {
		tsconfig.CompilerOptions.Target = target
	}
	if utsc.CompilerOptions.Lib == nil {
		tsconfig.CompilerOptions.Lib = []string{target, "dom"}
	}

	content, err := json.MarshalIndent(tsconfig, "", "\t")
	if err != nil {
		return nil, errors.Wrap(err, "marshaling tsconfig")
	}

	logger.Debug("Generated tsconfig.json: %s", strings.TrimSpace(string(content)))

	return content, nil
}

//go:embed node-shim.ts
var nodeShim string

func NodeShim(entrypoint string) (string, error) {
	// Remove the `.ts` suffix if one exists, since tsc doesn't accept
	// import paths with `.ts` endings. `.js` endings are fine.
	entrypoint = strings.TrimSuffix(entrypoint, ".ts")
	// The shim is stored under the .airplane directory.
	entrypoint = filepath.Join("../", entrypoint)
	// Escape for embedding into a string
	entrypoint = backslashEscape(entrypoint, `"`)

	shim, err := applyTemplate(nodeShim, struct {
		Entrypoint string
	}{
		Entrypoint: entrypoint,
	})
	if err != nil {
		return "", errors.Wrap(err, "templating shim")
	}

	return shim, nil
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

	return applyTemplate(heredoc.Doc(`
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
	`), struct {
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
