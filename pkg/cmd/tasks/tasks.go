package tasks

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/tasks/create"
	"github.com/airplanedev/cli/pkg/cmd/tasks/execute"
	"github.com/airplanedev/cli/pkg/cmd/tasks/list"
	"github.com/airplanedev/cli/pkg/cmd/tasks/push"
	"github.com/airplanedev/cli/pkg/cmd/tasks/show"
	"github.com/spf13/cobra"
)

// New returns a new cobra command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
		Long:  "Manage tasks",
		Example: heredoc.Doc(`
			$ airplane tasks create -f mytask.yml
			$ airplane tasks execute my-task
			$ airplane tasks push my-task -f mytask.yml
		`),
	}

	cmd.AddCommand(create.New(c))
	cmd.AddCommand(push.New(c))
	cmd.AddCommand(list.New(c))
	cmd.AddCommand(execute.New(c))
	cmd.AddCommand(show.New(c))

	return cmd
}
