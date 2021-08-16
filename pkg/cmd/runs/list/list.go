package list

import (
	"context"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	slug  string
	limit int
	since utils.TimeValue
	until utils.TimeValue
}

// New returns a new list command.
func New(c *cli.Config) *cobra.Command {
	var cfg config

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists runs",
		Example: heredoc.Doc(`
			airplane runs list
			airplane runs list --task <slug>
			airplane runs list --task <slug> -o json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c, cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.slug, "task", "t", "", "Filter runs by task slug")
	cmd.Flags().IntVar(&cfg.limit, "limit", 100, "If >0, returns at most --limit items.")
	cmd.Flags().Var(&cfg.since, "since", "Include only runs created after the given time")
	cmd.Flags().Var(&cfg.until, "until", "Include only runs created before the given time")

	return cmd
}

// Run runs the list command.
func run(ctx context.Context, c *cli.Config, cfg config) error {
	var client = c.Client

	req := api.ListRunsRequest{
		Limit: cfg.limit,
		Since: time.Time(cfg.since),
		Until: time.Time(cfg.until),
	}

	// If a task slug was provided, look up its task ID:
	if cfg.slug != "" {
		task, err := client.GetTask(ctx, cfg.slug)
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
