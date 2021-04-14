package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new create command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [<key_name>]",
		Short: "Generates a new API key for self-hosting agents and building custom integrations",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return run(cmd.Root().Context(), c, name)
		},
	}
	return cmd
}

// Run runs the create command.
func run(ctx context.Context, c *cli.Config, name string) error {
	var client = c.Client

	if name == "" {
		name = fmt.Sprintf("API Key (created %s)", time.Now().Format(time.RFC3339))
	}

	req := api.CreateAPIKeyRequest{
		Name: name,
	}
	logger.Log("  Creating API key named %s...", logger.Blue(req.Name))
	resp, err := client.CreateAPIKey(ctx, req)
	if err != nil {
		return errors.Wrap(err, "creating API key")
	}

	apiKey := resp.APIKey
	logger.Log("  Done!")
	if isatty.IsTerminal(os.Stdout.Fd()) {
		logger.Log(`  Save this somewhere safe, as you won't be able to retrieve it later:

  %s

  Key ID: %s
  Team ID: %s
`, logger.Blue(apiKey.Key), apiKey.ID, apiKey.TeamID)
	} else {
		if err := json.NewEncoder(os.Stdout).Encode(apiKey); err != nil {
			return errors.Wrap(err, "encoding API key to JSON")
		}
	}

	return nil
}
