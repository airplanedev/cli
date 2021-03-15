package root

import (
	"errors"

	"github.com/airplanedev/cli/commands/create"
	"github.com/airplanedev/cli/commands/execute"
	"github.com/airplanedev/cli/commands/list"
	"github.com/airplanedev/cli/commands/login"
	"github.com/airplanedev/cli/commands/push"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/spf13/cobra"
)

// New returns a new root cobra command.
func New() *cobra.Command {
	var output string
	var cfg = &cli.Config{
		Client: &api.Client{},
	}

	cmd := &cobra.Command{
		Use:   "airplane <command>",
		Short: "Airplane CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if c, err := conf.ReadDefault(); err == nil {
				cfg.Client.Token = c.Token
			}

			switch output {
			case "json":
				print.DefaultFormatter = print.JSON{}
			case "yaml":
				print.DefaultFormatter = print.YAML{}
			case "table":
				print.DefaultFormatter = print.Table{}
			default:
				return errors.New("--output must be (json|yaml|table)")
			}

			return nil
		},
	}

	// Silence usage and errors.
	//
	// Allows us to control how the output looks like.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	// Persistent flags, set globally to all commands.
	cmd.PersistentFlags().StringVarP(&cfg.Client.Host, "host", "", api.Host, "Airplane API Host.")
	cmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "The format to use for output (json|yaml|table).")

	// Most used sub commands.
	cmd.AddCommand(login.New(cfg))
	cmd.AddCommand(create.New(cfg))
	cmd.AddCommand(list.New(cfg))
	cmd.AddCommand(execute.New(cfg))
	cmd.AddCommand(push.New(cfg))

	return cmd
}
