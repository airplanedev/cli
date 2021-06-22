package build

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

// Golang creates a dockerfile for Go.
func golang(root string, options api.KindOptions) (string, error) {
	gomod := filepath.Join(root, "go.mod")
	gosum := filepath.Join(root, "go.sum")
	entrypoint, _ := options["entrypoint"].(string)
	main := filepath.Join(root, entrypoint)

	if err := fsx.AssertExistsAll(gomod, main); err != nil {
		return "", err
	}

	t, err := template.New("golang").Parse(`
FROM {{ .Base }} as builder

WORKDIR /airplane

COPY go.mod {{ if .HasGoSum -}} go.sum {{ end -}} ./
RUN go mod download

COPY . .

RUN ["go", "build", "-o", "/bin/main", "{{ .Entrypoint }}"]

FROM {{ .Base }}

COPY --from=builder /bin/main /bin/main

ENTRYPOINT ["/bin/main"]
`)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	v, err := GetVersion(NameGo, "1")
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, struct {
		Base       string
		Entrypoint string
		HasGoSum   bool
	}{
		Base:       v.String(),
		Entrypoint: filepath.Join("/airplane", entrypoint),
		HasGoSum:   fsx.AssertExistsAll(gosum) == nil,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}
