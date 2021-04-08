package configs

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/cmd/configs/get"
	"github.com/airplanedev/cli/pkg/cmd/configs/set"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configs",
		Short:   "Manage config variables",
		Long:    "Manage config variables",
		Aliases: []string{"config"},
		Example: heredoc.Doc(`
			$ airplane configs set my_database_url postgresql://my_database
			$ airplane configs get my_config_name
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(context.TODO(), c)
		}),
	}

	cmd.AddCommand(set.New(c))
	cmd.AddCommand(get.New(c))

	return cmd
}
