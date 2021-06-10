package definitions

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
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

func (err errReadDefinition) Error() string {
	return err.msg
}

// Implements ErrorExplained
func (err errReadDefinition) ExplainError() string {
	msgs := []string{}
	msgs = append(msgs, err.errorMsgs...)
	if len(err.errorMsgs) > 0 {
		msgs = append(msgs, "")
	}
	msgs = append(msgs, fmt.Sprintf("For more information on the task definition format, see the docs:\n%s", taskDefDocURL))
	return strings.Join(msgs, "\n")
}
