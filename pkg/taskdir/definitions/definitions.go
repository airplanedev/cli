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

func (this Definition) FromKindAndOptions(kind string, options api.KindOptions) error {
	if kind == "deno" {
		deno := DenoDefinition{
			Entrypoint: options["entrypoint"],
		}
		this.Deno = &deno

	} else if kind == "docker" {
		docker := DockerDefinition{
			Dockerfile: options["dockerfile"],
		}
		this.Docker = &docker

	} else if kind == "go" {
		godef := GoDefinition{
			Entrypoint: options["entrypoint"],
		}
		this.Go = &godef

	} else if kind == "node" {
		node := NodeDefinition{
			Entrypoint:  options["entrypoint"],
			Language:    options["language"],
			NodeVersion: options["nodeVersion"],
		}
		this.Node = &node

	} else if kind == "python" {
		python := PythonDefinition{
			Entrypoint: options["entrypoint"],
		}
		this.Python = &python

	} else if kind == "" {
		manual := ManualDefinition{
			Config: options,
		}
		this.Manual = &manual

	} else {
		return errors.Errorf("unknown kind specified: %s", kind)
	}

	return nil
}

func (this Definition) GetKindAndOptions() (string, api.KindOptions, error) {
	if this.Deno != nil {
		return "deno", api.KindOptions{
			"entrypoint": this.Deno.Entrypoint,
		}, nil
	} else if this.Docker != nil {
		return "docker", api.KindOptions{
			"dockerfile": this.Docker.Dockerfile,
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
		return "", api.KindOptions(this.Manual.Config), nil
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
