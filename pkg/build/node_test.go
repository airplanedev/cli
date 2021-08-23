package build

import (
	"encoding/json"
	"testing"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/stretchr/testify/require"
)

// -(map[string]interface {}) {
// 	+(map[string]interface {}) (len=2) {
// 	+ (string) (len=15) "compilerOptions": (map[string]interface {}) (len=8) {
// 	+  (string) (len=7) "allowJs": (bool) true,
// 	+  (string) (len=15) "esModuleInterop": (bool) true,
// 	+  (string) (len=3) "lib": ([]interface {}) (len=2) {
// 	+   (string) (len=6) "es2020",
// 	+   (string) (len=3) "dom"
// 	+  },
// 	+  (string) (len=6) "module": (string) (len=8) "commonjs",
// 	+  (string) (len=6) "outDir": (string) (len=6) "./dist",
// 	+  (string) (len=7) "rootDir": (string) (len=2) "..",
// 	+  (string) (len=12) "skipLibCheck": (bool) true,
// 	+  (string) (len=6) "target": (string) (len=6) "es2020"
// 	+ },
// 	+ (string) (len=5) "files": ([]interface {}) (len=1) {
// 	+  (string) (len=9) "./shim.ts"
// 	+ }
// 	 }

func TestGenTSConfig(t *testing.T) {
	require := require.New(t)

	// No tsconfig.json: should use defaults:
	c, err := GenTSConfig("testdata/tsconfigs/none", "testdata/tsconfigs/none/main.ts", nil)
	require.NoError(err)
	m := map[string]interface{}{}
	require.NoError(json.Unmarshal(c, &m))
	require.Equal(map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"allowJs":         true,
			"esModuleInterop": true,
			"lib":             []interface{}{"es2020", "dom"},
			"module":          "commonjs",
			"outDir":          "./dist",
			"rootDir":         "..",
			"skipLibCheck":    true,
			"target":          "es2020",
		},
		"files": []interface{}{"./shim.ts"},
	}, m)

	// Empty user-provided tsconfig.json: should use defaults:
	c, err = GenTSConfig("testdata/tsconfigs/empty", "testdata/tsconfigs/empty/main.ts", nil)
	require.NoError(err)
	m = map[string]interface{}{}
	require.NoError(json.Unmarshal(c, &m))
	require.Equal(map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"allowJs":         true,
			"esModuleInterop": true,
			"lib":             []interface{}{"es2020", "dom"},
			"module":          "commonjs",
			"outDir":          "./dist",
			"rootDir":         "..",
			"skipLibCheck":    true,
			"target":          "es2020",
		},
		"files": []interface{}{"./shim.ts"},
	}, m)

	// Empty user-provided tsconfig.json w/ node 12: should use older lib:
	c, err = GenTSConfig("testdata/tsconfigs/empty", "testdata/tsconfigs/empty/main.ts", api.KindOptions{
		"nodeVersion": "12",
	})
	require.NoError(err)
	m = map[string]interface{}{}
	require.NoError(json.Unmarshal(c, &m))
	require.Equal(map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"allowJs":         true,
			"esModuleInterop": true,
			"lib":             []interface{}{"es2019", "dom"}, // <-
			"module":          "commonjs",
			"outDir":          "./dist",
			"rootDir":         "..",
			"skipLibCheck":    true,
			"target":          "es2019", // <-
		},
		"files": []interface{}{"./shim.ts"},
	}, m)

	// Partially filled user-provided tsconfig.json: should accept overrides:
	c, err = GenTSConfig("testdata/tsconfigs/partial", "testdata/tsconfigs/partial/main.ts", nil)
	require.NoError(err)
	m = map[string]interface{}{}
	require.NoError(json.Unmarshal(c, &m))
	require.Equal(map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"allowJs": true,
			// esModuleInterop omitted
			// lib omitted
			"module":       "commonjs",
			"outDir":       "./dist",
			"rootDir":      "..",
			"skipLibCheck": true,
			// target omitted
		},
		"files": []interface{}{"./shim.ts"},
	}, m)

	// Fully specified user-provided tsconfig.json: should accept all:
	c, err = GenTSConfig("testdata/tsconfigs/full", "testdata/tsconfigs/full/main.ts", nil)
	require.NoError(err)
	m = map[string]interface{}{}
	require.NoError(json.Unmarshal(c, &m))
	require.Equal(map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			// allowJs omitted
			// esModuleInterop omitted
			// lib omitted
			// module omitted
			"outDir":  "./dist",
			"rootDir": "..",
			// skipLibCheck omitted
			// target omitted
		},
		"files": []interface{}{"./shim.ts"},
	}, m)
}
