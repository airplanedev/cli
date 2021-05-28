package definitions

import (
	"github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type ErrInvalidYAML struct {
	Msg string
}

func (err ErrInvalidYAML) Error() string {
	return err.Msg
}

type ErrSchemaValidation struct {
	Errors []gojsonschema.ResultError
}

func (err ErrSchemaValidation) Error() string {
	return "invalid YAML format"
}

// ValidateYAML checks that YAML data matches the schema defined by schemaObj
// Returns ErrInvalidYAML if not valid YAML and ErrSchemaValidation if YAML doesn't match schemaObj
func validateYAML(data []byte, schemaObj interface{}) error {
	var obj interface{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return errors.WithStack(ErrInvalidYAML{Msg: err.Error()})
	}

	r := &jsonschema.Reflector{PreferYAMLSchema: true}
	schemaLoader := gojsonschema.NewGoLoader(r.Reflect(schemaObj))
	docLoader := gojsonschema.NewGoLoader(obj)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return errors.Wrap(err, "validating schema")
	}

	if !result.Valid() {
		return errors.WithStack(ErrSchemaValidation{Errors: result.Errors()})
	}

	return nil
}
