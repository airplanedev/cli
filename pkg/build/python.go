package build

import (
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

// Python creates a dockerfile for Python.
func python(root string, args Args) (string, error) {
	var main = filepath.Join(root, args.entrypoint("main.py"))
	var reqs = filepath.Join(root, "requirements.txt")

	if err := exist(reqs, main); err != nil {
		return "", err
	}

	t, err := template.New("python").Parse(`
    FROM python:3.9.1-buster
    WORKDIR /airplane
    COPY . .
    RUN pip install -r requirements.txt
    ENTRYPOINT ["python", "{{ . }}"]
	`)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, path.Base(main)); err != nil {
		return "", err
	}

	return buf.String(), nil
}
