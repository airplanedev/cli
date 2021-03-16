package list

import (
	"context"
	"fmt"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new list command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists all tasks",
		Example: "airplane list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c)
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
		fmt.Printf("\n  There are no tasks yet, create a task:\n")
		fmt.Printf("\n    $ airplane create -f echo.yml\n")
		return nil
	}

	print.Tasks(res.Tasks)
	return nil
}
