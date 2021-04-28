package definitions

import (
	"fmt"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Definition represents a YAML-based task definition that can be used to create
// or update Airplane tasks.
//
// Note this is the subset of fields that can be represented with a revision,
// and therefore isolated to a specific environment.
type Definition Definition_0_2

func NewDefinitionFromTask(task api.Task) (Definition, error) {
	def := Definition{
		Slug:             task.Slug,
		Name:             task.Name,
		Description:      task.Description,
		Arguments:        task.Arguments,
		Parameters:       task.Parameters,
		Constraints:      task.Constraints,
		Env:              task.Env,
		ResourceRequests: task.ResourceRequests,
		Repo:             task.Repo,
		Timeout:          task.Timeout,
	}

	if task.Kind == "deno" {
		deno := DenoDefinition{
			Entrypoint: task.KindOptions["entrypoint"],
		}
		def.Deno = &deno

	} else if task.Kind == "docker" {
		docker := DockerDefinition{
			Dockerfile: task.KindOptions["dockerfile"],
		}
		def.Dockerfile = &docker

	} else if task.Kind == "go" {
		godef := GoDefinition{
			Entrypoint: task.KindOptions["entrypoint"],
		}
		def.Go = &godef

	} else if task.Kind == "node" {
		node := NodeDefinition{
			Entrypoint:  task.KindOptions["entrypoint"],
			Language:    task.KindOptions["language"],
			NodeVersion: task.KindOptions["nodeVersion"],
		}
		def.Node = &node

	} else if task.Kind == "python" {
		python := PythonDefinition{
			Entrypoint: task.KindOptions["entrypoint"],
		}
		def.Python = &python

	} else if task.Kind == "" {
		manual := ManualDefinition{
			Image:   task.Image,
			Command: task.Command,
		}
		def.Manual = &manual

	} else {
		return Definition{}, errors.Errorf("unknown kind specified: %s", task.Kind)
	}

	return def, nil
}

func (this Definition) GetKindAndOptions() (string, api.KindOptions, error) {
	if this.Deno != nil {
		return "deno", api.KindOptions{
			"entrypoint": this.Deno.Entrypoint,
		}, nil
	} else if this.Dockerfile != nil {
		return "docker", api.KindOptions{
			"dockerfile": this.Dockerfile.Dockerfile,
		}, nil
	} else if this.Go != nil {
		return "go", api.KindOptions{
			"entrypoint": this.Go.Entrypoint,
		}, nil
	} else if this.Node != nil {
		return "node", api.KindOptions{
			"entrypoint":  this.Node.Entrypoint,
			"language":    this.Node.Language,
			"nodeVersion": this.Node.NodeVersion,
		}, nil
	} else if this.Python != nil {
		return "python", api.KindOptions{
			"entrypoint": this.Python.Entrypoint,
		}, nil
	} else if this.Manual != nil {
		return "", api.KindOptions{}, nil
	}

	return "", api.KindOptions{}, errors.New("No kind specified")
}

func (this Definition) Validate() (Definition, error) {
	if this.Slug == "" {
		return this, errors.New("Expected a task slug")
	}

	// TODO: validate the rest of the fields!

	return this, nil
}

func UnmarshalDefinition(buf []byte, defPath string) (Definition, error) {
	// Validate definition against our Definition struct
	if err := validateYAML(buf, Definition{}); err != nil {
		// Try older definitions?
		if def, oerr := tryOlderDefinitions(buf); oerr == nil {
			return def, nil
		}

		// Print any "expected" validation errors
		switch err := errors.Cause(err).(type) {
		case ErrInvalidYAML:
			return Definition{}, newErrReadDefinition(fmt.Sprintf("Error reading %s: invalid YAML", defPath))
		case ErrSchemaValidation:
			errorMsgs := []string{}
			for _, verr := range err.Errors {
				errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", verr.Field(), verr.Description()))
			}
			return Definition{}, newErrReadDefinition(fmt.Sprintf("Error reading %s", defPath), errorMsgs...)
		default:
			return Definition{}, errors.Wrapf(err, "reading %s", defPath)
		}
	}

	var def Definition
	if err := yaml.Unmarshal(buf, &def); err != nil {
		return Definition{}, errors.Wrap(err, "unmarshalling task definition")
	}

	return def, nil
}

func tryOlderDefinitions(buf []byte) (Definition, error) {
	var err error
	if err = validateYAML(buf, Definition_0_1{}); err == nil {
		var def Definition_0_1
		if e := yaml.Unmarshal(buf, &def); e != nil {
			return Definition{}, err
		}
		return def.upgrade()
	}
	return Definition{}, err
}
