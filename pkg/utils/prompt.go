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

func PickSlug(def string, opts ...survey.AskOpt) (string, error) {
	opts = append(opts,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
		// TODO: add a validator to ensure this slug is unique.
		survey.WithValidator(func(val interface{}) error {
			if str, ok := val.(string); !ok || !IsSlug(str) {
				return errors.New("Slugs can only contain lowercase letters, underscores, and numbers.")
			}

			return nil
		}),
	)

	var slug string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Pick a unique identifier (slug) for this task",
			Default: def,
		},
		&slug,
		opts...,
	); err != nil {
		return "", errors.Wrap(err, "prompting for slug")
	}

	return slug, nil
}
