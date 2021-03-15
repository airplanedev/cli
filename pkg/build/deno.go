package build

import (
	"html/template"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Deno creates a dockerfile for Deno.
func deno(root string, args Args) (string, error) {
	var main = filepath.Join(root, args.entrypoint("main.ts"))

	if err := exist(main); err != nil {
		return "", err
	}

	t, err := template.New("deno").Parse(`
FROM hayd/alpine-deno:1.7.2
WORKDIR /airplane
ADD . .
RUN deno cache {{ . }}
USER deno
ENTRYPOINT ["deno", "run", "-A", "{{ . }}"]
	`)
	if err != nil {
		return "", errors.Wrap(err, "new template")
	}

	var buf strings.Builder
	if err := t.Execute(&buf, path.Base(main)); err != nil {
		return "", err
	}

	return buf.String(), nil
}
