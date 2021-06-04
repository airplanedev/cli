package build

import (
	"path/filepath"
	"strings"
	"text/template"
)

// Python creates a dockerfile for Python.
func python(root string, args Args) (string, error) {
	var entrypoint = args["entrypoint"]
	var main = filepath.Join(root, entrypoint)
	var reqs = filepath.Join(root, "requirements.txt")

	if err := exist(main); err != nil {
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
		HasRequirements: exist(reqs) == nil,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
