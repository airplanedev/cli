package get

import (
	"context"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/configs"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new get command.
func New(c *cli.Config) *cobra.Command {
	var secret bool
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get a config variable's value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c, args[0])
		},
	}
	cmd.Flags().BoolVar(&secret, "secret", false, "Whether to set config var as a secret")
	return cmd
}

// Run runs the get command.
func run(ctx context.Context, c *cli.Config, name string) error {
	var client = c.Client

	nt, err := configs.ParseName(name)
	if err == configs.ErrInvalidConfigName {
		return errors.Errorf("invalid config name: %s - expected my_config or my_config:tag", name)
	}
	resp, err := client.GetConfig(ctx, nt.Name, nt.Tag)
	if err != nil {
		return errors.Wrap(err, "get config")
	}

	print.Config(resp.Config)
	return nil
}
