package apikeys

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/apikeys/create"
	"github.com/airplanedev/cli/pkg/cmd/apikeys/delete"
	"github.com/airplanedev/cli/pkg/cmd/apikeys/list"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

// New returns a new cobra command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apikeys",
		Short:   "Manage API keys",
		Long:    "Manage API keys",
		Aliases: []string{"apikey"},
		Example: heredoc.Doc(`
			airplane apikeys create "Agent Key"
			airplane apikeys list
			airplane apikeys delete <key_id>
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
	}

	cmd.AddCommand(create.New(c))
	cmd.AddCommand(delete.New(c))
	cmd.AddCommand(list.New(c))

	return cmd
}
