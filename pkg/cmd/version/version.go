package version

import (
	"context"
	"fmt"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/version"
	"github.com/spf13/cobra"
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
	return fmt.Sprintf("Version: %s (%s)", version.Get(), version.Date())
}

func run(ctx context.Context) error {
	logger.Log(Version())

	if err := version.CheckLatest(ctx); err != nil {
		return err
	}

	return nil
}
