package python

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPythonInstalled(t *testing.T) {
	require := require.New(t)

	// Assumes python3 is installed in test environment...
	err := checkPythonInstalled(context.Background())
	require.NoError(err)
}
