package utils

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

func Confirm(question string) (bool, error) {
	ok := false
	if err := survey.AskOne(
		&survey.Confirm{
			Message: question,
		},
		&ok,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return false, errors.Wrap(err, "confirming")
	}

	return ok, nil
}

func CanPrompt() bool {
	return isatty.IsTerminal(os.Stderr.Fd())
}
