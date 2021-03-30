package build

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeVersion(t *testing.T) {
	t.Run("expand versions", func(t *testing.T) {
		var assert = require.New(t)
		assert.Equal("15.12", expandNodeVersion(""))
		assert.Equal("15.12", expandNodeVersion("15"))
		assert.Equal("14.16", expandNodeVersion("14"))
		assert.Equal("12.22", expandNodeVersion("12"))
		assert.Equal("16.23-alpha", expandNodeVersion("16.23-alpha"))
	})
}
