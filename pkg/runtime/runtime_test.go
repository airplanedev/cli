package runtime

import (
	"errors"
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

			v, err := Pathof(path, filename)

			assert.Nil(err)
			assert.Equal("with_package_json", filepath.Base(v))
		})

		t.Run("task within monorepo", func(t *testing.T) {
			var assert = require.New(t)
			var path = "testdata/monorepo/my/task"
			var filename = "package.json"

			v, err := Pathof(path, filename)

			assert.Nil(err)
			assert.Equal("monorepo", filepath.Base(v))
		})

		t.Run("missing package.json", func(t *testing.T) {
			var assert = require.New(t)
			var path = "testdata"
			var filename = "package.json"

			v, err := Pathof(path, filename)

			assert.Error(err)
			assert.True(errors.Is(err, ErrMissing), "expected a runtime.ErrMissing")
			assert.Equal("", v)
		})
	})
}
