package list

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
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
		Short: "Lists runs",
		Example: heredoc.Doc(`
			airplane runs list
			airplane runs list --task <slug>
			airplane runs list --task <slug> -o json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c, slug)
		},
	}

	cmd.Flags().StringVarP(&slug, "task", "t", "", "Filter runs by task slug")

	return cmd
}

// Run runs the list command.
func run(ctx context.Context, c *cli.Config, slug string) error {
	var client = c.Client

	req := api.ListRunsRequest{}

	// If a task slug was provided, look up its task ID:
	if slug != "" {
		task, err := client.GetTask(ctx, slug)
		if err != nil {
			return err
		}
		req.TaskID = task.ID
	}

	resp, err := client.ListRuns(ctx, req)
	if err != nil {
		return errors.Wrap(err, "list runs")
	}

	print.Runs(resp.Runs)
	return nil
}
