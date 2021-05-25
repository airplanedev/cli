package javascript

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
	runtime.Register(".js", Runtime{})
}

// Code template.
var code = template.Must(template.New("js").Parse(`// airplane: {{ .URL }}

export default async function(args){
  console.log('arguments: ', args);
}
`))

// Data represents the data template.
type data struct {
	URL string
}

// Runtime implementaton.
type Runtime struct{}

// Generate implementation.
func (r Runtime) Generate(t api.Task) ([]byte, error) {
	var args = data{URL: t.URL}
	var buf bytes.Buffer

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("javascript: template execute - %w", err)
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
