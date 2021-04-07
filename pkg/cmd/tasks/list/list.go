package list

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new list command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all tasks",
		Example: heredoc.Doc(`
			airplane tasks list
			airplane tasks list -o json
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c)
		},
	}
	return cmd
}

// Run runs the list command.
func run(ctx context.Context, c *cli.Config) error {
	var client = c.Client

	res, err := client.ListTasks(ctx)
	if err != nil {
		return errors.Wrap(err, "list tasks")
	}

	if len(res.Tasks) == 0 {
		logger.Log(`
  There are no tasks yet. To create a sample task:
    airplane tasks deploy -f github.com/airplanedev/examples/node/hello-world-javascript/airplane.yml`)
		return nil
	}

	print.Tasks(res.Tasks)
	return nil
}
