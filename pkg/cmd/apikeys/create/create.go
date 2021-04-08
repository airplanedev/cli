package create

import (
	"context"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new create command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generates a new API key for self-hosting agents and building custom integrations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c)
		},
	}
	return cmd
}

// Run runs the create command.
func run(ctx context.Context, c *cli.Config) error {
	var client = c.Client

	resp, err := client.CreateAPIKey(ctx)
	if err != nil {
		return errors.Wrap(err, "creating API key")
	}

	print.APIKeyCreated(resp.APIKey)
	return nil
}
