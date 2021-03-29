package initcmd

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func initFromScratch(cmd *cobra.Command, cfg config) error {
	runtime, err := pickRuntime()
	if err != nil {
		return err
	}

	name, err := pickString("Pick a name:")
	if err != nil {
		return err
	}

	description, err := pickString("Pick a description:")
	if err != nil {
		return err
	}

	file := cfg.file
	if file == "" {
		file = "airplane.yml"
	}

	dir, err := taskdir.Open(file)
	if err != nil {
		return err
	}
	defer dir.Close()

	def := taskdir.Definition{
		// TODO: choose a unique slug via the Airplane API
		Slug:        utils.MakeSlug(name),
		Name:        name,
		Description: description,
	}

	if runtime == runtimeKindManual {
		// TODO: select an image + entrypoint
	} else {
		def.Builder, def.BuilderConfig = defaultRuntimeConfig(runtime)
	}

	if err := dir.WriteDefinition(def); err != nil {
		return err
	}

	// TODO: maybe ask for a parameter?

	cmd.Printf(`An Airplane task definition for '%s' has been created in %s!

(TODO: instructions)

Then, deploy it to Airplane with:
	airplane tasks deploy -f %s`, name, file, file)

	return nil
}

func defaultRuntimeConfig(runtime runtimeKind) (string, api.BuilderConfig) {
	switch runtime {
	case runtimeKindDeno:
		return "deno", api.BuilderConfig{
			"entrypoint": "main.ts",
		}
	case runtimeKindDockerfile:
		return "docker", api.BuilderConfig{
			"dockerfile": "Dockerfile",
		}
	case runtimeKindGo:
		return "docker", api.BuilderConfig{
			"dockerfile": "Dockerfile",
		}
	case runtimeKindManual:
		return "docker", api.BuilderConfig{
			"dockerfile": "Dockerfile",
		}
	case runtimeKindNode:
		return "docker", api.BuilderConfig{
			"dockerfile": "Dockerfile",
		}
	case runtimeKindPython:
		return "docker", api.BuilderConfig{
			"dockerfile": "Dockerfile",
		}
	default:
		return "", nil
	}
}

type runtimeKind string

var (
	runtimeKindNode       runtimeKind = "Node.js"
	runtimeKindPython     runtimeKind = "Python"
	runtimeKindDeno       runtimeKind = "Deno"
	runtimeKindDockerfile runtimeKind = "Dockerfile"
	runtimeKindGo         runtimeKind = "Go"
	runtimeKindManual     runtimeKind = "Pre-built Docker image"
)

func pickRuntime() (runtimeKind, error) {
	var runtime string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Pick a runtime:",
			Options: []string{
				string(runtimeKindNode),
				string(runtimeKindPython),
				string(runtimeKindDeno),
				string(runtimeKindDockerfile),
				string(runtimeKindGo),
				string(runtimeKindManual),
			},
			Default: string(runtimeKindNode),
		},
		&runtime,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return runtimeKind(""), errors.Wrap(err, "selecting runtime")
	}

	return runtimeKind(runtime), nil
}

func pickString(msg string) (string, error) {
	var str string
	if err := survey.AskOne(
		&survey.Input{
			Message: msg,
		},
		&str,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return "", errors.Wrap(err, "")
	}

	return str, nil
}
