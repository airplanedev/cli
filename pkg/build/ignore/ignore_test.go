package ignore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerignorePatterns(tt *testing.T) {
	for _, test := range []struct {
		In  string
		Out string
	}{
		{"", ""},
		{"node_modules", "**/node_modules"},
		{"!node_modules", "!**/node_modules"},
		{"/node_modules", "node_modules"},
		{"!/node_modules", "!node_modules"},
	} {
		tt.Run(test.In, func(t *testing.T) {
			require.Equal(t, test.Out, toDockerignore(test.In))
		})
	}
}
