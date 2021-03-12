package main

import (
	"context"
	"fmt"
	"os"

	"github.com/airplanedev/cli/commands/root"
	"github.com/airplanedev/cli/pkg/trap"
	"github.com/pkg/errors"
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
		if errors.Is(err, context.Canceled) {
			// TODO(amir): output operation canceled?
			return
		}
		fmt.Println("")
		fmt.Println("  Error: ", errors.Cause(err).Error())
		fmt.Println("")
		os.Exit(1)
	}
}
