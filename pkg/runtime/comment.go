package runtime

import (
	"net/url"
	"path"
	"regexp"

	"github.com/airplanedev/cli/pkg/api"
)

var (
	// commentRegex matches against the string produced by Comment() below.
	//
	// It is used to extract a slug from a comment in a script file.
	commentRegex = regexp.MustCompile(`Linked to (https://.*air.*/t/.*) \[do not edit this line\]`)
)

// Comment generates a linking comment that is used
// to associate a script file with an Airplane task.
//
// This comment can be parsed out of a script file using Slug.
func Comment(r Interface, task api.Task) string {
	return r.FormatComment("Linked to " + task.URL + " [do not edit this line]")
}

func Slug(code []byte) (string, bool) {
	result := commentRegex.FindSubmatch(code)
	if len(result) == 0 {
		return "", false
	}

	u, err := url.Parse(string(result[1]))
	if err != nil {
		return "", false
	}

	_, slug := path.Split(u.Path)

	return slug, slug != ""
}
