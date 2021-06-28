package build

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInlineString(t *testing.T) {
	require := require.New(t)

	require.Equal(
		`echo 'The sheep couldn'"'"'t sleep, no matter how many humans he counted.'`,
		inlineString(`The sheep couldn't sleep, no matter how many humans he counted.`),
	)
	require.Equal(
		`echo ''"'"''"'"''"'"''`,
		inlineString(`'''`),
	)
}
