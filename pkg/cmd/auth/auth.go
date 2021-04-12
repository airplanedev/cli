package auth

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/info"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/cmd/auth/logout"
	"github.com/spf13/cobra"
)

func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage CLI authentication",
		Long:  "Manage CLI authentication",
		Example: heredoc.Doc(`
			$ airplane auth login
			$ airplane auth logout
		`),
	}

	cmd.AddCommand(info.New(c))
	cmd.AddCommand(login.New(c))
	cmd.AddCommand(logout.New(c))

	return cmd
}
