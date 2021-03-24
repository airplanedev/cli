package tasks

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/tasks/deploy"
	"github.com/airplanedev/cli/pkg/cmd/tasks/execute"
	"github.com/airplanedev/cli/pkg/cmd/tasks/get"
	"github.com/airplanedev/cli/pkg/cmd/tasks/list"
	"github.com/spf13/cobra"
)

// New returns a new cobra command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
		Long:  "Manage tasks",
		Example: heredoc.Doc(`
			$ airplane tasks deploy my-task -f mytask.yml
			$ airplane tasks get my-task
			$ airplane tasks execute my-task
		`),
	}

	cmd.AddCommand(deploy.New(c))
	cmd.AddCommand(list.New(c))
	cmd.AddCommand(execute.New(c))
	cmd.AddCommand(get.New(c))

	return cmd
}
