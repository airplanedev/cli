package initcmd

import (
	"encoding/json"
	"path"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
)

type runtimeScaffolder interface {
	GenerateFiles(def taskdir.Definition, filemap map[string][]byte) error
}

// Default noop scaffolder

type noopScaffolder struct{}

var _ runtimeScaffolder = noopScaffolder{}

func (this noopScaffolder) GenerateFiles(def taskdir.Definition, filemap map[string][]byte) error {
	return nil
}

// Deno

type denoScaffolder struct {
	entrypoint string
}

var _ runtimeScaffolder = denoScaffolder{}

func (this denoScaffolder) GenerateFiles(def taskdir.Definition, filemap map[string][]byte) error {
	// entrypoint
	filemap[path.Join(def.Root, this.entrypoint)] = []byte(`console.log("Hello world!");
`)
	return nil
}

// Node

type nodeScaffolder struct {
	entrypoint string
}

var _ runtimeScaffolder = nodeScaffolder{}

func (this nodeScaffolder) GenerateFiles(def taskdir.Definition, filemap map[string][]byte) error {
	// entrypoint
	filemap[path.Join(def.Root, this.entrypoint)] = []byte(`const main = (args) => {
	console.log("Hello world!")
}

main(process.argv.slice(2));
`)

	// package.json
	j, err := json.MarshalIndent(struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}{
		Name:    def.Slug,
		Version: "0.0.1",
	}, "", "  ")
	if err != nil {
		return errors.Wrap(err, "creating package.json")
	}
	filemap[path.Join(filepath.Dir(def.Root), "package.json")] = j
	return nil
}
