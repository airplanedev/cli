package django

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	managepy = `#!/usr/bin/env python3
print("task_2")
print("task_1")`
	boom = `#!/usr/bin/env python3
	\\
	`
)

func TestFramework(t *testing.T) {
	t.Run("missing manage.py", func(t *testing.T) {
		var assert = require.New(t)
		var root = tmpdir(t)

		f, err := Open(root)

		assert.Error(err)
		assert.Nil(f)
	})

	t.Run("has manage.py", func(t *testing.T) {
		var assert = require.New(t)
		var root = tmpdir(t)

		err := ioutil.WriteFile(
			filepath.Join(root, "manage.py"),
			[]byte(managepy),
			0777,
		)
		assert.NoError(err)

		f, err := Open(root)

		assert.NoError(err)
		assert.NotNil(f)
	})

	t.Run("list commands", func(t *testing.T) {
		var assert = require.New(t)
		var root = tmpdir(t)
		var bin = filepath.Join(root, "manage.py")

		err := os.WriteFile(
			bin,
			[]byte(managepy),
			0770,
		)
		assert.NoError(err)

		f, err := Open(root)
		assert.NoError(err)

		cmds, err := f.ListCommands()
		assert.NoError(err)
		assert.Equal([]string{"task_1", "task_2"}, cmds)
	})

	t.Run("list command error", func(t *testing.T) {
		var assert = require.New(t)
		var root = tmpdir(t)
		var bin = filepath.Join(root, "manage.py")

		err := os.WriteFile(
			bin,
			[]byte(boom),
			0770,
		)
		assert.NoError(err)

		f, err := Open(root)
		assert.NoError(err)

		_, err = f.ListCommands()
		assert.Error(err)
		assert.Contains(err.Error(), "unexpected character")
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
