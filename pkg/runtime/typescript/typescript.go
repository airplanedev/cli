package typescript

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/runtime/javascript"
)

// Init register the runtime.
func init() {
	runtime.Register(".ts", Runtime{})
}

// Code template.
var code = template.Must(template.New("ts").Parse(`{{.Comment}}

type Params = {
  {{- range .Params }}
  {{ .Name }}: {{ .Type }}
  {{- end }}
}

export default async function(params: Params){
  console.log('parameters:', params);
}
`))

// Data represents the data template.
type data struct {
	Comment string
	Params  []param
}

// Param represents the parameter.
type param struct {
	Name string
	Type string
}

// Runtime implementaton.
type Runtime struct {
	javascript.Runtime
}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{Comment: runtime.Comment(t)}
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
	case api.TypeUpload:
		return "string"
	default:
		return "unknown"
	}
}
