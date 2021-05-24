package typescript

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/scaffold"
)

// Init register the generator.
func init() {
	scaffold.Register(".ts", Generator{})
}

// Code template.
var code = template.Must(template.New("ts").Parse(`
// airplane: {{ .Slug }}

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
	Slug   string
	Params []param
}

// Param represents the parameter.
type param struct {
	Name string
	Type string
}

// Generator implementaton.
type Generator struct{}

// Generate implementation.
func (g Generator) Generate(t api.Task) ([]byte, error) {
	var args = data{Slug: t.Slug}
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
	default:
		// TODO(amir): ideally how do we convert upload this type in typescript.
		return "unknown"
	}
}
