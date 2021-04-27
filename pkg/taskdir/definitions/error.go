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
