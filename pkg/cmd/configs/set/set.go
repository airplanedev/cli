package set

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/configs"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
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
		value, err = readValue(secret)
		if err != nil {
			return err
		}
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

func readValue(secret bool) (string, error) {
	var value string
	if utils.CanPrompt() {
		msg := "Config value:"
		var prompt survey.Prompt
		if secret {
			prompt = &survey.Password{Message: msg}
		} else {
			prompt = &survey.Input{Message: msg}
		}
		if err := survey.AskOne(
			prompt,
			&value,
			survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
		); err != nil {
			return "", errors.Wrap(err, "prompting value")
		}
	} else {
		// Read from stdin
		logger.Log("Reading secret from stdin...")
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", errors.Wrap(err, "reading from stdin")
		}
		value = string(data)
	}
	return strings.TrimSpace(value), nil
}
