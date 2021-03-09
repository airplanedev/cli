package root

import (
	"github.com/airplanedev/cli/commands/create"
	"github.com/airplanedev/cli/commands/list"
	"github.com/airplanedev/cli/commands/login"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/spf13/cobra"
)

// New returns a new root cobra command.
func New() *cobra.Command {
	var cfg = &cli.Config{
		Client: &api.Client{},
	}

	cmd := &cobra.Command{
		Use:   "airplane <command>",
		Short: "Airplane CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if c, err := conf.ReadDefault(); err == nil {
				cfg.Client.Token = c.Token
			}
		},
	}

	// Persistent flags, set globally to all commands.
	cmd.PersistentFlags().StringVarP(&cfg.Client.Host, "host", "", api.Host, "Airplane API Host.")

	// Most used sub commands.
	cmd.AddCommand(login.New(cfg))
	cmd.AddCommand(create.New(cfg))
	cmd.AddCommand(list.New(cfg))

	return cmd
}
