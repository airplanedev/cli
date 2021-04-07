package initcmd

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
)

func initFromScratch(cfg config) error {
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

	dir, err := taskdir.New(file)
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
		// TODO: let folks enter an image
		def.Image = "alpine:3"
		def.Command = []string{"echo", `"Hello World"`}
	} else {
		if def.Builder, def.BuilderConfig, err = defaultRuntimeConfig(runtime); err != nil {
			return err
		}
	}

	if err := dir.WriteDefinition(def); err != nil {
		return err
	}

	logger.Log(`
A skeleton Airplane task definition for '%s' has been created in %s! Fill it out with the rest of your task details.

Once you are ready, deploy it to Airplane with:
  airplane tasks deploy -f %s`, name, file, file)

	return nil
}

func defaultRuntimeConfig(runtime runtimeKind) (string, api.BuilderConfig, error) {
	// TODO: let folks configure the following configuration
	switch runtime {
	case runtimeKindDeno:
		return "deno", api.BuilderConfig{
			"entrypoint": "main.ts",
		}, nil
	case runtimeKindDockerfile:
		return "docker", api.BuilderConfig{
			"dockerfile": "Dockerfile",
		}, nil
	case runtimeKindGo:
		return "go", api.BuilderConfig{
			"entrypoint": "main.go",
		}, nil
	case runtimeKindNode:
		return "node", api.BuilderConfig{
			"entrypoint":  "main.js",
			"language":    "javascript",
			"nodeVersion": "15",
		}, nil
	case runtimeKindPython:
		return "python", api.BuilderConfig{
			"entrypoint": "main.py",
		}, nil
	default:
		return "", nil, errors.Errorf("unknown runtime: %s", runtime)
	}
}

type runtimeKind string

const (
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
		return "", errors.Wrap(err, "prompting")
	}

	return str, nil
}
