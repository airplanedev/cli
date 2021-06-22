package build

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

// Deno creates a dockerfile for Deno.
func deno(root string, options api.KindOptions) (string, error) {
	entrypoint, _ := options["entrypoint"].(string)
	main := filepath.Join(root, entrypoint)

	if err := fsx.AssertExistsAll(main); err != nil {
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

	v, err := GetVersion(NameDeno, "1")
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
