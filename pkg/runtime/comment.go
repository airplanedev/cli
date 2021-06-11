package runtime

import (
	"fmt"
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

// ErrNotLinked is an error that is raised when a path unexpectedly
// does not contain a slug. It can be used to explain to a user how
// they should link that file with a task.
type ErrNotLinked struct {
	Path string
}

func (e ErrNotLinked) Error() string {
	return fmt.Sprintf(
		"the file %s is not linked to a task",
		e.Path,
	)
}

func (e ErrNotLinked) ExplainError() string {
	return fmt.Sprintf(
		"You can link the file by running:\n  airplane init --slug <slug> %s",
		e.Path,
	)
}
