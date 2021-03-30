package utils

import (
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
)

func init() {
	slug.MaxLength = 50
}

// Make returns a slug generated from the provided string.
func MakeSlug(s string) string {
	// We prefer underscores over dashes since they are easier
	// to use in Go templates.
	return strings.ReplaceAll(slug.Make(s), "-", "_")
}

// IsSlug returns True if the provided text does not contain whitespace
// characters, punctuation, uppercase letters, and non-ASCII characters.
// It can contain `_`, but not at the beginning or end of the text.
// It should be of length <= to MaxLength.
// All output from MakeSlug(text) will pass this test.
func IsSlug(text string) bool {
	// The slug library will accept text with `-`'s, so we need to add our own check.
	return slug.IsSlug(text) && !strings.Contains(text, "-")
}

func PickSlug(def string) (string, error) {
	var slug string
	if err := survey.AskOne(
		&survey.Input{
			Message: "Pick a unique identifier (slug) for this task:",
			Default: def,
		},
		&slug,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
		// TODO: add a validator to ensure this slug is unique.
		survey.WithValidator(func(val interface{}) error {
			if str, ok := val.(string); !ok || !IsSlug(str) {
				return errors.New("Slugs can only contain lowercase letters, underscores, and numbers.")
			}

			return nil
		}),
	); err != nil {
		return "", errors.Wrap(err, "prompting for slug")
	}

	return slug, nil
}
