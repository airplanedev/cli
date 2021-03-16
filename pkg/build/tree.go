package build

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
	"github.com/pkg/errors"
)

// Tree represents the build tree.
//
// It uses docker's archive library to create
// a tarball and pass it onto the build operation.
type Tree struct {
	root string
}

// NewTree returns a new tree in a temporary directory.
func NewTree() (*Tree, error) {
	tmpdir, err := ioutil.TempDir("", "airplane_context_*")
	if err != nil {
		return nil, errors.Wrap(err, "tempdir")
	}
	return &Tree{root: tmpdir}, nil
}

// Copy copies src into the tree.
func (t *Tree) Copy(src string) error {
	r, err := archive.TarWithOptions(src, &archive.TarOptions{
		Compression: archive.Uncompressed,
	})
	if err != nil {
		return errors.Wrap(err, "tar with options")
	}

	if err := archive.Unpack(r, t.root, &archive.TarOptions{}); err != nil {
		return errors.Wrapf(err, "unpack %s", t.root)
	}

	return nil
}

// Write writes the given r into dst.
func (t *Tree) Write(dst string, r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "read")
	}

	return ioutil.WriteFile(filepath.Join(t.root, dst), buf, 0600)
}

// Archive archives the tree and returns a tarball.
func (t *Tree) Archive() (io.ReadCloser, error) {
	r, err := archive.Tar(t.root, archive.Gzip)
	if err != nil {
		return nil, errors.Wrap(err, "tar")
	}
	return r, nil
}

// Close discards the tree.
func (t *Tree) Close() error {
	return os.RemoveAll(t.root)
}
