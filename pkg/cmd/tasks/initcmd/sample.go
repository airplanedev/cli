package initcmd

import (
	"os"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
)

func initFromSample(cfg config) error {
	runtime, err := pickRuntime()
	if err != nil {
		return err
	}

	samplepath, err := pickSample(runtime)
	if err != nil {
		return err
	}

	dir, err := taskdir.Open(samplepath)
	if err != nil {
		return err
	}
	defer dir.Close()

	def, err := dir.ReadDefinition()
	if err != nil {
		return err
	}

	var outputdir string
	if cfg.file != "" {
		// If a user passed a file path, store the sample in the user-provided directory.
		// In the case of `-f airplane.yml`, that would be the current directory.
		outputdir = path.Dir(cfg.file)
	} else {
		// By default, store the sample in a directory with the same name
		// as the containing directory in GitHub.
		outputdir = path.Base(path.Dir(dir.DefinitionPath()))
	}

	// Rename the sample definition to match the filename in the `-f` argument.
	// This is done to maintain semantic consistency with the other kinds of
	// init, but is not strictly necessary.
	defname := path.Base(dir.DefinitionPath())
	if cfg.file != "" && defname != path.Base(cfg.file) {
		defname = path.Base(cfg.file)
		if err := os.Rename(
			dir.DefinitionPath(),
			path.Join(path.Dir(dir.DefinitionPath()), defname),
		); err != nil {
			return errors.Wrap(err, "renaming task definitino")
		}
	}

	// Copy the sample code from the temporary directory into the user's
	// local directory.
	if err := copy.Copy(dir.Dir, outputdir); err != nil {
		return errors.Wrap(err, "copying sample directory")
	}

	file := path.Join(outputdir, defname)
	logger.Log(`
An Airplane task definition for '%s' has been created!

To deploy it to Airplane, run:
	airplane tasks deploy -f %s`, def.Name, file)

	return nil
}

func pickSample(runtime runtimeKind) (string, error) {
	// This maps runtimes to a list of allowlisted examples where the
	// key is the label shown to users in the select menu and the value
	// is the file path (using `airplane tasks deploy -f` semantics) of
	// that example's task definition.
	//
	// For simplicity, we explicitly manage this list here rather
	// than dynamically fetching it from GitHub. However, we should
	// eventually make this dynamic so that old versions of the CLI
	// do not break if/when we change the layout of the examples repo.
	//
	// Feel free to allowlist more examples here as they are
	// added upstream.
	samplesByRuntime := map[runtimeKind]map[string]string{
		runtimeKindDeno: {
			"Hello World": "github.com/airplanedev/examples/deno/hello-world/airplane.yml",
		},
		runtimeKindDockerfile: {
			"Hello World": "github.com/airplanedev/examples/docker/hello-world/airplane.yml",
		},
		runtimeKindGo: {
			"Hello World": "github.com/airplanedev/examples/go/hello-world/airplane.yml",
		},
		runtimeKindManual: {
			"Hello World": "github.com/airplanedev/examples/manual/hello-world/airplane.yml",
			"Print Env":   "github.com/airplanedev/examples/manual/env/airplane.yml",
		},
		runtimeKindNode: {
			"Hello World":              "github.com/airplanedev/examples/node/hello-world-javascript/airplane.yml",
			"Hello World (TypeScript)": "github.com/airplanedev/examples/node/hello-world-typescript/airplane.yml",
		},
		runtimeKindPython: {
			"Hello World": "github.com/airplanedev/examples/python/hello-world/airplane.yml",
		},
	}
	samples, ok := samplesByRuntime[runtime]
	if !ok {
		return "", errors.Errorf("Unexpected runtime: %s", runtime)
	}
	options := []string{}
	for label := range samples {
		options = append(options, label)
	}

	var selected string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Pick a sample:",
			Options: options,
		},
		&selected,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return "", errors.Wrap(err, "selecting sample")
	}

	sample, ok := samplesByRuntime[runtime][selected]
	if !ok {
		return "", errors.Errorf("Unexpected sample selected; %s", selected)
	}

	return sample, nil
}
