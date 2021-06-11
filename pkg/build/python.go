package build

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
)

// Python creates a dockerfile for Python.
func python(root string, options api.KindOptions) (string, error) {
	entrypoint, _ := options["entrypoint"].(string)
	main := filepath.Join(root, entrypoint)
	reqs := filepath.Join(root, "requirements.txt")

	if err := fsx.AssertExistsAll(main); err != nil {
		return "", err
	}

	t, err := template.New("python").Parse(`
    FROM {{ .Base }}
    WORKDIR /airplane
		{{if not .HasRequirements}}
		RUN echo > requirements.txt
		{{end}}
    COPY . .
    RUN pip install -r requirements.txt
    ENTRYPOINT ["python", "/airplane/{{ .Entrypoint }}"]
	`)
	if err != nil {
		return "", err
	}

	v, err := GetVersion(NamePython, "3")
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, struct {
		Base            string
		Entrypoint      string
		HasRequirements bool
	}{
		Base:            v.String(),
		Entrypoint:      entrypoint,
		HasRequirements: fsx.AssertExistsAll(reqs) == nil,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
