package build

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// Deno creates a dockerfile for Deno.
func deno(root string, args Args) (string, error) {
	var entrypoint = args["entrypoint"]
	var main = filepath.Join(root, entrypoint)

	if err := exist(main); err != nil {
		return "", err
	}

	t, err := template.New("deno").Parse(`
FROM {{ .Base }}
WORKDIR /airplane
ADD . .
RUN deno cache {{ .Entrypoint }}
USER deno
ENTRYPOINT ["deno", "run", "-A", "{{ .Entrypoint }}"]
	`)
	if err != nil {
		return "", errors.Wrap(err, "new template")
	}

	v, err := GetVersion(BuilderNameDeno, "1")
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, struct {
		Base       string
		Entrypoint string
	}{
		Base:       v.String(),
		Entrypoint: entrypoint,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
