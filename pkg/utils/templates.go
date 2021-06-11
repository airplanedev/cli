package utils

import (
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// ApplyTemplate executes template t with the provided data and
// returns the output.
func ApplyTemplate(t string, data interface{}) (string, error) {
	tmpl, err := template.New("airplane").Parse(t)
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", errors.Wrap(err, "executing template")
	}

	return buf.String(), nil
}
