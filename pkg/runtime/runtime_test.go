package runtime

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntime(t *testing.T) {
	t.Run("pathof", func(t *testing.T) {
		t.Run("task with package.json", func(t *testing.T) {
			var assert = require.New(t)
			var path = "testdata/my/task/with_package_json"
			var filename = "package.json"

			v, ok := Pathof(path, filename)

			assert.True(ok)
			assert.Equal("with_package_json", filepath.Base(v))
		})

		t.Run("task within monorepo", func(t *testing.T) {
			var assert = require.New(t)
			var path = "testdata/monorepo/my/task"
			var filename = "package.json"

			v, ok := Pathof(path, filename)

			assert.True(ok)
			assert.Equal("monorepo", filepath.Base(v))
		})

		t.Run("missing package.json", func(t *testing.T) {
			var assert = require.New(t)
			var path = "testdata/monorepo/my/task"
			var filename = "package.json"

			v, ok := Pathof(path, filename)

			assert.True(ok)
			assert.Equal("monorepo", filepath.Base(v))
		})
	})
}
