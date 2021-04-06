package set

import (
	"context"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/configs"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	blue  = color.New(color.FgHiBlue).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
)

// New returns a new set command.
func New(c *cli.Config) *cobra.Command {
	var secret bool
	cmd := &cobra.Command{
		Use:   "set [--secret] <name> <value>",
		Short: "Set a new or existing config variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, args[0], args[1], secret)
		},
	}
	cmd.Flags().BoolVar(&secret, "secret", false, "Whether to set config var as a secret")
	return cmd
}

// Run runs the set command.
func run(ctx context.Context, c *cli.Config, name, value string, secret bool) error {
	var client = c.Client

	nt, err := configs.ParseName(name)
	if err == configs.ErrInvalidConfigName {
		return errors.Errorf("invalid config name: %s - expected my_config or my_config:tag", name)
	} else if err != nil {
		return errors.Wrap(err, "parsing config name")
	}
	// Avoid printing back secrets
	var valueStr string
	if secret {
		valueStr = "<secret value>"
	} else {
		valueStr = value
	}
	logger.Log("  Setting %s to %s...", blue(name), green(valueStr))
	req := api.SetConfigRequest{
		Name:     nt.Name,
		Tag:      nt.Tag,
		Value:    value,
		IsSecret: secret,
	}
	if err := client.SetConfig(ctx, req); err != nil {
		return errors.Wrap(err, "set config")
	}
	logger.Log("  Done!")
	return nil
}
