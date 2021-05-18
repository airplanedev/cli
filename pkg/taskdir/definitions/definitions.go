package definitions

import (
	"fmt"
	"strings"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/mitchellh/mapstructure"
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

	if task.Kind == api.TaskKindDeno {
		def.Deno = &DenoDefinition{}
		if err := mapstructure.Decode(task.KindOptions, &def.Deno); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Deno options")
		}

	} else if task.Kind == api.TaskKindDocker {
		def.Dockerfile = &DockerDefinition{}
		if err := mapstructure.Decode(task.KindOptions, &def.Dockerfile); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Dockerfile options")
		}

	} else if task.Kind == api.TaskKindGo {
		def.Go = &GoDefinition{}
		if err := mapstructure.Decode(task.KindOptions, &def.Go); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Go options")
		}

	} else if task.Kind == api.TaskKindNode {
		def.Node = &NodeDefinition{}
		if err := mapstructure.Decode(task.KindOptions, &def.Node); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Node options")
		}

	} else if task.Kind == api.TaskKindPython {
		def.Python = &PythonDefinition{}
		if err := mapstructure.Decode(task.KindOptions, &def.Python); err != nil {
			return Definition{}, errors.Wrap(err, "decoding Python options")
		}

	} else if task.Kind == api.TaskKindManual {
		def.Manual = &ManualDefinition{
			Image:   task.Image,
			Command: task.Command,
		}

	} else if task.Kind == api.TaskKindSQL {
		def.SQL = &SQLDefinition{}
		if err := mapstructure.Decode(task.KindOptions, &def.SQL); err != nil {
			return Definition{}, errors.Wrap(err, "decoding SQL options")
		}

	} else {
		return Definition{}, errors.Errorf("unknown kind specified: %s", task.Kind)
	}

	return def, nil
}

func (this Definition) GetKindAndOptions() (api.TaskKind, api.KindOptions, error) {
	options := api.KindOptions{}
	if this.Deno != nil {
		if err := mapstructure.Decode(this.Deno, &options); err != nil {
			return "", api.KindOptions{}, errors.Wrap(err, "decoding Deno definition")
		}
		return api.TaskKindDeno, options, nil
	} else if this.Dockerfile != nil {
		if err := mapstructure.Decode(this.Dockerfile, &options); err != nil {
			return "", api.KindOptions{}, errors.Wrap(err, "decoding Dockerfile definition")
		}
		return api.TaskKindDocker, options, nil
	} else if this.Go != nil {
		if err := mapstructure.Decode(this.Go, &options); err != nil {
			return "", api.KindOptions{}, errors.Wrap(err, "decoding Go definition")
		}
		return api.TaskKindGo, options, nil
	} else if this.Node != nil {
		if err := mapstructure.Decode(this.Node, &options); err != nil {
			return "", api.KindOptions{}, errors.Wrap(err, "decoding Node definition")
		}
		return api.TaskKindNode, options, nil
	} else if this.Python != nil {
		if err := mapstructure.Decode(this.Python, &options); err != nil {
			return "", api.KindOptions{}, errors.Wrap(err, "decoding Python definition")
		}
		return api.TaskKindPython, options, nil
	} else if this.Manual != nil {
		return api.TaskKindManual, api.KindOptions{}, nil
	} else if this.SQL != nil {
		if err := mapstructure.Decode(this.SQL, &options); err != nil {
			return "", api.KindOptions{}, errors.Wrap(err, "decoding SQL definition")
		}
		return api.TaskKindSQL, options, nil
	}

	return "", api.KindOptions{}, errors.New("No kind specified")
}

func (this Definition) Validate() (Definition, error) {
	if this.Slug == "" {
		return this, errors.New("Expected a task slug")
	}

	defs := []string{}
	if this.Manual != nil {
		defs = append(defs, "manual")
	}
	if this.Deno != nil {
		defs = append(defs, "deno")
	}
	if this.Dockerfile != nil {
		defs = append(defs, "dockerfile")
	}
	if this.Go != nil {
		defs = append(defs, "go")
	}
	if this.Node != nil {
		defs = append(defs, "node")
	}
	if this.Python != nil {
		defs = append(defs, "python")
	}
	if this.SQL != nil {
		defs = append(defs, "sql")
	}

	if len(defs) == 0 {
		return this, errors.New("No task type defined")
	}
	if len(defs) > 1 {
		return this, errors.Errorf("Too many task types defined: only one of (%s) expected", strings.Join(defs, ", "))
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
			return Definition{}, newErrReadDefinition(fmt.Sprintf("Error reading %s, invalid YAML:\n  %s", defPath, err))
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
