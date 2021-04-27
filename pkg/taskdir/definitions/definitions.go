package definitions

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Definition represents a YAML-based task definition that can be used to create
// or update Airplane tasks.
//
// Note this is the subset of fields that can be represented with a revision,
// and therefore isolated to a specific environment.
type Definition Definition_0_2

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
