package build

import (
	"context"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/pkg/errors"
)

// Request represents a build request.
type Request struct {
	Local   bool
	Client  *api.Client
	Root    string
	Def     definitions.Definition
	TaskID  string
	TaskEnv api.TaskEnv
	Shim    bool
}

// Response represents a build response.
type Response struct {
	ImageURL string
}

// Run runs the build and returns an image URL.
func Run(ctx context.Context, req Request) (*Response, error) {
	if req.Local {
		return local(ctx, req)
	}
	return remote(ctx, req)
}

// applyTemplate executes template t with the provided data and
// returns the output.
func applyTemplate(t string, data interface{}) (string, error) {
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

func inlineString(s string) string {
	// To inline a multi-line string into a Dockerfile, insert `\n\` characters:
	s = strings.Join(strings.Split(s, "\n"), "\\n\\\n")
	// Since the string is wrapped in single-quotes, escape any single-quotes
	// inside of the target string.
	s = strings.ReplaceAll(s, "'", `'"'"'`)
	return "echo '" + s + "'"
}
