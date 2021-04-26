package taskdir

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const taskDefDocURL = "https://docs.airplane.dev/reference/task-definition-reference"

type errReadDefinition struct {
	msg       string
	errorMsgs []string
}

func newErrReadDefinition(msg string, errorMsgs ...string) error {
	return errors.WithStack(errReadDefinition{
		msg:       msg,
		errorMsgs: errorMsgs,
	})
}

func (this errReadDefinition) Error() string {
	return this.msg
}

// Implements ErrorExplained
func (this errReadDefinition) ExplainError() string {
	msgs := []string{}
	msgs = append(msgs, this.errorMsgs...)
	if len(this.errorMsgs) > 0 {
		msgs = append(msgs, "")
	}
	msgs = append(msgs, fmt.Sprintf("For more information on the task definition format, see the docs:\n%s", taskDefDocURL))
	return strings.Join(msgs, "\n")
}

func (this TaskDirectory) ReadDefinition() (Definition, error) {
	buf, err := ioutil.ReadFile(this.defPath)
	if err != nil {
		return Definition{}, errors.Wrap(err, "reading task definition")
	}

	// Validate definition against our Definition struct
	if err := ValidateYAML(buf, Definition{}); err != nil {
		// Try older definitions?
		if def, oerr := tryOlderDefinitions(buf); oerr == nil {
			return def, nil
		}

		defPath := this.defPath
		// Attempt to set a prettier defPath, best effort
		if wd, err := os.Getwd(); err != nil {
			logger.Debug("%s", err)
		} else if path, err := filepath.Rel(wd, defPath); err != nil {
			logger.Debug("%s", err)
		} else {
			defPath = path
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
	if err = ValidateYAML(buf, Definition_0_1{}); err == nil {
		var def Definition_0_1
		if e := yaml.Unmarshal(buf, &def); e != nil {
			return Definition{}, err
		}
		return def.upgrade()
	}
	return Definition{}, err
}

// WriteSlug updates the slug of a task definition and persists this to disk.
//
// It attempts to retain the existing file's formatting (comments, etc.) where possible.
func (this TaskDirectory) WriteSlug(slug string) error {
	if err := utils.SetYAMLField(this.defPath, "slug", slug); err != nil {
		return errors.Wrap(err, "setting slug")
	}

	return nil
}

func (this TaskDirectory) WriteDefinition(def Definition) error {
	data, err := yaml.Marshal(def)
	if err != nil {
		return errors.Wrap(err, "marshalling definition")
	}

	if err := ioutil.WriteFile(this.defPath, data, 0664); err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}
