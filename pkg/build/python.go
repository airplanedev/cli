package build

import (
	_ "embed"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

// Python shim code.
//go:embed shim.py
var pythonShim string

// Python creates a dockerfile for Python.
func python(root string, args api.KindOptions) (string, error) {
	entrypoint, _ := args["entrypoint"].(string)
	main := filepath.Join(root, entrypoint)
	init := filepath.Join(root, "__init__.py")
	reqs := filepath.Join(root, "requirements.txt")

	if args["shim"] != "true" {
		return pythonLegacy(root, args)
	}

	if err := fsx.AssertExistsAll(main); err != nil {
		return "", err
	}

	v, err := GetVersion(NamePython, "3")
	if err != nil {
		return "", err
	}

	shim, err := applyTemplate(pythonShim, struct {
		Entrypoint string
	}{
		Entrypoint: filepath.Join("/airplane", entrypoint),
	})
	if err != nil {
		return "", errors.Wrapf(err, "rendering shim")
	}

	const dockerfile = `
    FROM {{ .Base }}
    WORKDIR /airplane
    RUN echo '{{.Shim}}' > /shim.py
    {{if not .HasInit}}
    RUN touch __init__.py
    {{end}}
    COPY . .
		{{if .HasRequirements}}
    RUN pip install -r requirements.txt
		{{end}}
    ENTRYPOINT ["python", "/shim.py", "/airplane/{{ .Entrypoint }}"]
	`

	df, err := applyTemplate(dockerfile, struct {
		Base            string
		Shim            string
		Entrypoint      string
		HasRequirements bool
		HasInit         bool
	}{
		Base:            v.String(),
		Shim:            strings.Join(strings.Split(shim, "\n"), "\\n\\\n"),
		Entrypoint:      entrypoint,
		HasRequirements: fsx.Exists(reqs),
		HasInit:         fsx.Exists(init),
	})
	if err != nil {
		return "", errors.Wrapf(err, "rendering dockerfile")
	}

	return df, nil
}

// PythonLegacy generates a dockerfile for legacy python support.
func pythonLegacy(root string, args api.KindOptions) (string, error) {
	var entrypoint, _ = args["entrypoint"].(string)
	var main = filepath.Join(root, entrypoint)
	var reqs = filepath.Join(root, "requirements.txt")

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
