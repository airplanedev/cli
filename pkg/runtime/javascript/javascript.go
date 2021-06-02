package javascript

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strings"
	"text/template"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/runtime"
)

// Init register the runtime.
func init() {
	runtime.Register(".js", Runtime{})
}

// CommentPrefix.
const (
	commentPrefix = "// Linked to Airplane task, do not edit:"
)

// Code template.
var code = template.Must(template.New("js").Parse(`{{.Comment}}

export default async function(args){
  console.log('arguments: ', args);
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
	var args = data{Comment: r.Comment(t)}
	var buf bytes.Buffer

	if err := code.Execute(&buf, args); err != nil {
		return nil, fmt.Errorf("javascript: template execute - %w", err)
	}

	return buf.Bytes(), nil
}

// Slug implementation.
func (r Runtime) Slug(code []byte) (string, bool) {
	var s = bufio.NewScanner(bytes.NewReader(code))

	for s.Scan() {
		var line = strings.TrimSpace(s.Text())

		if strings.HasPrefix(line, commentPrefix) {
			continue
		}

		rawurl := strings.TrimSpace(strings.TrimPrefix(line, "//"))
		u, err := url.Parse(rawurl)
		if err != nil {
			return "", false
		}

		_, slug := path.Split(u.Path)
		return slug, len(slug) > 0
	}

	return "", false
}

// Comment implementation.
func (r Runtime) Comment(t api.Task) string {
	return fmt.Sprintf("%s\n// %s", commentPrefix, t.URL)
}
