package main

import (
	"os"

	"github.com/airplanedev/cli/commands/root"
	"github.com/airplanedev/cli/pkg/trap"
	_ "github.com/segmentio/events/v2/text"
)

var (
	version = "<dev>"
)

func main() {
	var cmd = root.New()
	var ctx = trap.Context()

	cmd.Version = version

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
