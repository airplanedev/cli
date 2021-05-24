package main

import (
	"context"
	"fmt"
	"os"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
)

// This is used to test the Node.js shim. It will be removed
// before merging and replaced by a future PR adding
// support for deploying via the standard `deploy` command.
func main() {
	if len(os.Args) != 3 {
		fmt.Printf("shimtest [filepath] [host]")
		os.Exit(1)
	}
	path := os.Args[1]
	host := os.Args[2]

	client := &api.Client{
		Host: host,
	}
	if c, err := conf.ReadDefault(); err != nil {
		fmt.Printf("client error: %+v", err)
	} else {
		client.Token = c.Tokens["api.airplane.so:5000"]
	}

	dir, err := taskdir.New(path)
	if err != nil {
		fmt.Printf("filepath error: %+v", err)
	}

	if resp, err := build.Run(context.Background(), build.Request{
		Builder: build.BuilderKindLocal,
		Client:  client,
		Dir:     dir,
		Def: definitions.Definition(definitions.Definition_0_2{
			Node: &definitions.NodeDefinition{
				Entrypoint: path,
			},
		}),
		TaskID:  "123",
		TaskEnv: api.TaskEnv{},
	}); err != nil {
		fmt.Printf("build error: %+v", err)
		os.Exit(1)
	} else {
		fmt.Printf("resp: %+v", resp)
	}
}
