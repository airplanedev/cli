package version

import (
	"context"
	"fmt"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/spf13/cobra"
)

// Set by Go Releaser.
var (
	version     string = "<unknown>"
	compileDate string = "<unknown>"
)

func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Long:  "Print the CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context())
		},
	}

	return cmd
}

func Version() string {
	return fmt.Sprintf("Version: %s (%s)", version, compileDate)
}

func run(ctx context.Context) error {
	logger.Log(Version())

	return nil
}
