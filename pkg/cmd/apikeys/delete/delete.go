package delete

import (
	"context"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	red = color.New(color.FgHiRed).SprintFunc()
)

// New returns a new delete command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <key_id>...",
		Short: "Deletes one or more API keys by ID",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c, args)
		},
	}
	return cmd
}

// Run runs the delete command.
func run(ctx context.Context, c *cli.Config, apiKeyIDs []string) error {
	var client = c.Client

	for _, apiKeyID := range apiKeyIDs {
		req := api.DeleteAPIKeyRequest{
			KeyID: apiKeyID,
		}
		logger.Log("  Deleting key %s...", red(apiKeyID))
		if err := client.DeleteAPIKey(ctx, req); err != nil {
			return errors.Wrap(err, "deleting API key")
		}
	}
	logger.Log("  Done.")
	return nil
}
