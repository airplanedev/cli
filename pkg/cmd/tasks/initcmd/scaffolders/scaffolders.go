package scaffolders

import (
	"encoding/json"
	"path"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/pkg/errors"
)

type RuntimeScaffolder interface {
	GenerateFiles(def definitions.Definition, filemap map[string][]byte) error
}

// Deno

type DenoScaffolder struct {
	Entrypoint string
}

var _ RuntimeScaffolder = DenoScaffolder{}

func (this DenoScaffolder) GenerateFiles(def definitions.Definition, filemap map[string][]byte) error {
	// Entrypoint
	filemap[path.Join(def.Root, this.Entrypoint)] = []byte(`console.log("Hello world!");
`)
	return nil
}

// Dockerfile

type DockerfileScaffolder struct {
	Dockerfile string
}

var _ RuntimeScaffolder = DockerfileScaffolder{}

func (this DockerfileScaffolder) GenerateFiles(def definitions.Definition, filemap map[string][]byte) error {
	// Dockerfile
	filemap[path.Join(def.Root, this.Dockerfile)] = []byte(`FROM ubuntu:20.10

CMD ["echo", "hello world"]
`)
	return nil
}

// Golang

type GoScaffolder struct {
	Entrypoint string
}

var _ RuntimeScaffolder = GoScaffolder{}

func (this GoScaffolder) GenerateFiles(def definitions.Definition, filemap map[string][]byte) error {
	// Entrypoint
	filemap[path.Join(def.Root, this.Entrypoint)] = []byte(`package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, playground")
}
`)
	// TODO: this should also add go.mod and go.sum for Go to be fully buildable
	return nil
}

// Node

type NodeScaffolder struct {
	Entrypoint string
}

var _ RuntimeScaffolder = NodeScaffolder{}

func (this NodeScaffolder) GenerateFiles(def definitions.Definition, filemap map[string][]byte) error {
	// Entrypoint
	filemap[path.Join(def.Root, this.Entrypoint)] = []byte(`const main = (args) => {
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

// Python

type PythonScaffolder struct {
	Entrypoint string
}

var _ RuntimeScaffolder = PythonScaffolder{}

func (this PythonScaffolder) GenerateFiles(def definitions.Definition, filemap map[string][]byte) error {
	// Entrypoint
	filemap[path.Join(def.Root, this.Entrypoint)] = []byte(`print("Hello world!")
`)
	// Requirements
	filemap[path.Join(def.Root, "requirements.txt")] = []byte("\n")
	return nil
}
