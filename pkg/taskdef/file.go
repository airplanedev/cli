package taskdef

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

type file struct {
	path string
}

var _ io.Reader = file{}

func (this file) Read(p []byte) (n int, err error) {
	f, err := os.Open(this.path)
	if err != nil {
		return 0, errors.Wrap(err, "opening file")
	}
	defer f.Close()

	return f.Read(p)
}
