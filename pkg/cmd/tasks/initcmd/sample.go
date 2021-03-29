package initcmd

import (
	"os"
	"path"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func initFromSample(cmd *cobra.Command, cfg config) error {
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

	var filedir string
	if cfg.file == "" {
		// By default, store the sample in a directory with the same name
		// as the containing directory in GitHub.
		filedir = path.Base(path.Dir(dir.DefinitionPath()))
	} else {
		// Otherwise, store it in the user-provided directory.
		// In the case of `-f airplane.yml`, that would be the current directory.
		filedir = path.Dir(cfg.file)
	}
	// TODO: consider renaming the task definition to match what a user provided with `-f`.
	file := path.Join(filedir, path.Base(dir.DefinitionPath()))

	if err := copy.Copy(dir.Dir, filedir); err != nil {
		return errors.Wrap(err, "copying sample directory")
	}

	cmd.Printf(`An Airplane task definition for '%s' has been created!

To deploy it to Airplane, run:
	airplane tasks deploy -f %s
`, def.Name, file)

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
			"Hello World": "github.com/airplanedev/examples/deno/hello-world/airplane.yml",
		},
		runtimeKindManual: {
			"Hello World": "github.com/airplanedev/examples/deno/hello-world/airplane.yml",
			"Print Env":   "github.com/airplanedev/examples/deno/env/airplane.yml",
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
