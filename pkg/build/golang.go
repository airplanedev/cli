package build

import (
	"html/template"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Golang creates a dockerfile for Go.
func golang(root string, args Args) (string, error) {
	var gomod = filepath.Join(root, "go.mod")
	var gosum = filepath.Join(root, "go.sum")
	var entrypoint = args["entrypoint"]
	var main = filepath.Join(root, entrypoint)

	if err := exist(gomod, main); err != nil {
		return "", err
	}

	t, err := template.New("golang").Parse(`
FROM golang:{{ .GoVersion }} as builder

WORKDIR /airplane

COPY go.mod {{ if .HasGoSum -}} go.sum {{ end -}} ./
RUN go mod download

COPY . .

RUN ["go", "build", "-o", "/bin/main", "/airplane/{{ .Main }}"]

FROM golang:{{ .GoVersion }}

COPY --from=builder /bin/main /bin/main

ENTRYPOINT ["/bin/main"]
`)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	data := struct {
		Main      string
		HasGoSum  bool
		GoVersion string
	}{
		Main:      entrypoint,
		HasGoSum:  exist(gosum) == nil,
		GoVersion: "1.16.3-alpine3.13",
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
