package list

import (
	"context"
	"fmt"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new list command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists all tasks",
		Example: "airplane list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}
	return cmd
}

// Run runs the list command.
func run(ctx context.Context) error {
	var client api.Client

	res, err := client.ListTasks(ctx)
	if err != nil {
		return errors.Wrap(err, "list tasks")
	}

	for _, t := range res.Tasks {
		fmt.Println(t.Name)
	}

	return nil
}
