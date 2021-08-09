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
	opts TreeOptions
}

type TreeOptions struct {
	// ExcludePatterns is a list of .dockerignore-style ignores.
	// These files will not be copied to the temporary directory.
	ExcludePatterns []string
}

// NewTree returns a new tree in a temporary directory.
func NewTree(opts TreeOptions) (*Tree, error) {
	tmpdir, err := ioutil.TempDir("", "airplane_context_*")
	if err != nil {
		return nil, errors.Wrap(err, "tempdir")
	}
	return &Tree{root: tmpdir, opts: opts}, nil
}

// Copy copies src into the tree.
func (t *Tree) Copy(src string) error {
	r, err := archive.TarWithOptions(src, &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: t.opts.ExcludePatterns,
	})
	if err != nil {
		return errors.Wrap(err, "tar with options")
	}

	if err := archive.Unpack(r, t.root, &archive.TarOptions{}); err != nil {
		return errors.Wrapf(err, "unpack %s", t.root)
	}

	return nil
}

// MkdirAll creates dir relative to root
func (t *Tree) MkdirAll(dir string) error {
	if err := os.MkdirAll(filepath.Join(t.root, dir), 0777); err != nil {
		return errors.Wrap(err, "making directory")
	}
	return nil
}

// Write writes the given r into dst.
func (t *Tree) Write(dst string, r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "write")
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
