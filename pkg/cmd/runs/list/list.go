package list

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new list command.
func New(c *cli.Config) *cobra.Command {
	var slug string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists runs for a task",
		Example: heredoc.Doc(`
			airplane runs list --task <slug>
			airplane runs list --task my-task -o json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, slug)
		},
	}

	cmd.Flags().StringVarP(&slug, "task", "t", "", "The task slug")
	cmd.MarkFlagRequired("task")

	return cmd
}

// Run runs the list command.
func run(ctx context.Context, c *cli.Config, slug string) error {
	var client = c.Client

	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return err
	}

	resp, err := client.ListRuns(ctx, task.ID)
	if err != nil {
		return errors.Wrap(err, "list runs")
	}

	print.Runs(resp.Runs)
	return nil
}
