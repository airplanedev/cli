package ap

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjects(t *testing.T) {
	t.Run("read missing", func(t *testing.T) {
		var assert = require.New(t)
		var root = tmpdir(t)

		_, err := ReadProject(root)

		assert.Error(err)
		assert.True(errors.Is(err, ErrNoProject))
	})

	t.Run("save, read", func(t *testing.T) {
		var assert = require.New(t)
		var root = tmpdir(t)
		var project = NewProject(root)

		project.Framework = "django"
		err := project.Save()
		assert.NoError(err)

		p, err := ReadProject(root)
		assert.NoError(err)
		assert.Equal(p, project)
	})
}

// Tmpdir creates a temporary directory for the duration
// of the given t.
func tmpdir(t testing.TB) string {
	t.Helper()

	path, err := ioutil.TempDir("", "airplane_")
	if err != nil {
		t.Fatalf("tmpdir: %s", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(path)
	})

	return path
}
