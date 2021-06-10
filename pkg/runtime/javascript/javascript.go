package javascript

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/pkg/errors"
)

// Init register the runtime.
func init() {
	runtime.Register(".js", Runtime{})
}

// Code template.
var code = template.Must(template.New("js").Parse(`// {{.Comment}}

export default async function(params){
  console.log('parameters:', params);
}
`))

// Data represents the data template.
type data struct {
	Comment string
}

// Runtime implementaton.
type Runtime struct{}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{Comment: runtime.Comment(t)}
	var buf bytes.Buffer

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("javascript: template execute - %w", err)
	}

	return buf.Bytes(), nil
}

// Workdir implementation.
func (r Runtime) Workdir(path string) (string, error) {
	return runtime.Pathof(path, "package.json")
}

// Root implementation.
//
// The method finds the nearest package.json, If the package.json contains
// any airplane settings with `root` definition it will use that as the root.
func (r Runtime) Root(path string) (string, error) {
	dst, err := runtime.Pathof(path, "package.json")
	if err != nil {
		return "", err
	}

	pkgjson := filepath.Join(dst, "package.json")
	buf, err := ioutil.ReadFile(pkgjson)
	if err != nil {
		return "", errors.Wrapf(err, "javascript: reading %s", dst)
	}

	var pkg struct {
		Settings runtime.Settings `json:"airplane"`
	}

	if err := json.Unmarshal(buf, &pkg); err != nil {
		return "", fmt.Errorf("javascript: reading %s - %w", dst, err)
	}

	if root := pkg.Settings.Root; root != "" {
		return filepath.Join(dst, root), nil
	}

	return dst, nil
}

// Kind implementation.
func (r Runtime) Kind() api.TaskKind {
	return api.TaskKindNode
}
