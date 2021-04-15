package set

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/configs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new set command.
func New(c *cli.Config) *cobra.Command {
	var secret bool
	cmd := &cobra.Command{
		Use:   "set [--secret] <name> [<value>]",
		Short: "Set a new or existing config variable",
		Example: heredoc.Doc(`
			# Pass in a value to the prompt
			$ airplane configs set --secret db/url
			Config value: my_value_here
			
			# Pass in a value by piping it in via stdin
			$ cat my_secret_value.txt | airplane configs set --secret secret_config

			# Recommended for non-secrets only - pass in a value via arguments
			$ airplane configs set nonsecret_config my_value
		`),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var value *string
			if len(args) == 2 {
				value = &args[1]
			}
			return run(cmd.Root().Context(), c, args[0], value, secret)
		},
	}
	cmd.Flags().BoolVar(&secret, "secret", false, "Whether to set config var as a secret")
	return cmd
}

// Run runs the set command.
func run(ctx context.Context, c *cli.Config, name string, argValue *string, secret bool) error {
	var client = c.Client

	nt, err := configs.ParseName(name)
	if err == configs.ErrInvalidConfigName {
		return errors.Errorf("invalid config name: %s - expected my_config or my_config:tag", name)
	} else if err != nil {
		return errors.Wrap(err, "parsing config name")
	}

	var value string
	if argValue != nil {
		value = *argValue
	} else {
		var err error
		value, err = configs.ReadValue(secret)
		if err != nil {
			return err
		}
	}
	return configs.SetConfig(ctx, client, nt, value, secret)
}
