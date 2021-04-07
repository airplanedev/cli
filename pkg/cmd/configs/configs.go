package configs

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/configs/get"
	"github.com/airplanedev/cli/pkg/cmd/configs/set"
	"github.com/spf13/cobra"
)

func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configs",
		Short: "Manage config variables",
		Long:  "Manage config variables",
		Example: heredoc.Doc(`
			$ airplane config set my_database_url postgresql://my_database
			$ airplane config get my_config_name
		`),
	}

	cmd.AddCommand(set.New(c))
	cmd.AddCommand(get.New(c))

	return cmd
}
