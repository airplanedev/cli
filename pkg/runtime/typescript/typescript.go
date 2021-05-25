package typescript

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/runtime"
)

// Init register the runtime.
func init() {
	runtime.Register(".ts", Runtime{})
}

// Code template.
var code = template.Must(template.New("ts").Parse(`// airplane: {{ .URL }}

type Params = {
  {{- range .Params }}
  {{ .Name }}: {{ .Type }}
  {{- end }}
}

export default async function(args: Params){
  console.log('arguments: ', args);
}
`))

// Data represents the data template.
type data struct {
	URL    string
	Params []param
}

// Param represents the parameter.
type param struct {
	Name string
	Type string
}

// Runtime implementaton.
type Runtime struct{}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{URL: t.URL}
	var params = t.Parameters
	var buf bytes.Buffer

	for _, p := range params {
		args.Params = append(args.Params, param{
			Name: p.Slug,
			Type: typeof(p.Type),
		})
	}

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("typescript: template execute - %w", err)
	}

	return buf.Bytes(), nil
}

// URL implementation.
func (r Runtime) URL(code []byte) (string, bool) {
	var s = bufio.NewScanner(bytes.NewReader(code))

	for s.Scan() {
		var line = strings.TrimSpace(s.Text())
		var parts = strings.Fields(line)
		var rawurl = parts[len(parts)-1]

		if !strings.HasPrefix(line, "// airplane:") {
			return "", false
		}

		u, err := url.Parse(rawurl)
		if err != nil {
			return "", false
		}

		return u.String(), true
	}

	return "", false
}

// Comment implementation.
func (r Runtime) Comment(t api.Task) string {
	return fmt.Sprintf("// airplane: %s", t.URL)
}

// Typeof translates the given type to typescript.
func typeof(t api.Type) string {
	switch t {
	case api.TypeInteger, api.TypeFloat:
		return "number"
	case api.TypeDate, api.TypeDatetime:
		return "string"
	case api.TypeBoolean:
		return "boolean"
	case api.TypeString:
		return "string"
	default:
		// TODO(amir): ideally how do we convert upload this type in typescript.
		return "unknown"
	}
}
