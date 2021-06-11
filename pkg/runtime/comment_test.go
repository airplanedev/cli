package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlug(tt *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		slug string
		ok   bool
	}{
		{
			name: "empty file",
			in:   ``,
			ok:   false,
		},
		{
			name: "missing comment",
			in:   `import airplane from 'airplane'`,
			ok:   false,
		},
		{
			name: "unrelated comment",
			in: `// Airplane (https://airplane.dev) is great!
console.log('ship it')`,
			ok: false,
		},
		{
			name: "extracts slug correctly",
			in: `// Linked to https://app.airplane.dev/t/myslug [do not edit this line]
console.log('ship it')`,
			slug: "myslug",
			ok:   true,
		},
		{
			name: "extracts slug correctly in staging",
			in: `// Linked to https://web.airstage.app/t/myslug [do not edit this line]
console.log('ship it')`,
			slug: "myslug",
			ok:   true,
		},
		{
			name: "extracts slug correctly in dev",
			in: `// Linked to https://app.airplane.so:5000/t/myslug [do not edit this line]
console.log('ship it')`,
			slug: "myslug",
			ok:   true,
		},
		{
			name: "extracts slug correctly regardless of line number",
			in: `import airplane from 'airplane'
console.log('ship it')

// Linked to https://app.airplane.dev/t/myslug [do not edit this line]`,
			slug: "myslug",
			ok:   true,
		},
	} {
		tt.Run(test.name, func(t *testing.T) {
			slug, ok := Slug([]byte(test.in))
			require.Equal(t, test.ok, ok)
			require.Equal(t, test.slug, slug)
		})
	}
}
