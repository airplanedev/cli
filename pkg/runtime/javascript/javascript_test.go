package javascript

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatComment(t *testing.T) {
	require := require.New(t)

	r := Runtime{}

	require.Equal("// test", r.FormatComment("test"))
	require.Equal(`// line 1
// line 2`, r.FormatComment(`line 1
line 2`))
}
