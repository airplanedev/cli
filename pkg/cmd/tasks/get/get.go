package get

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/spf13/cobra"
)

// New returns a new get command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get information about a task",
		Example: heredoc.Doc(`
			$ airplane tasks get my-task
			$ airplane tasks get my-task -o yaml
			$ airplane tasks get my-task -o json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, args[0])
		},
	}
	return cmd
}

// Run runs the get command.
func run(ctx context.Context, c *cli.Config, slug string) error {
	var client = c.Client

	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return err
	}

	print.Task(task)
	return nil
}
