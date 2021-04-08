package list

import (
	"context"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new list command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists API keys by ID and created time",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c)
		},
	}
	return cmd
}

// Run runs the list command.
func run(ctx context.Context, c *cli.Config) error {
	var client = c.Client

	resp, err := client.ListAPIKeys(ctx)
	if err != nil {
		return errors.Wrap(err, "creating API key")
	}

	print.APIKeys(resp.APIKeys)
	return nil
}
