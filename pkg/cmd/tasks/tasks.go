package tasks

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/cmd/tasks/deploy"
	"github.com/airplanedev/cli/pkg/cmd/tasks/execute"
	"github.com/airplanedev/cli/pkg/cmd/tasks/get"
	"github.com/airplanedev/cli/pkg/cmd/tasks/list"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

// New returns a new cobra command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
		Long:  "Manage tasks",
		Example: heredoc.Doc(`
			airplane tasks deploy -f mytask.yml
			airplane tasks get my_task
			airplane tasks execute my_task
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(context.TODO(), cmd, c)
		}),
	}

	cmd.AddCommand(deploy.New(c))
	cmd.AddCommand(list.New(c))
	cmd.AddCommand(execute.New(c))
	cmd.AddCommand(get.New(c))

	return cmd
}
