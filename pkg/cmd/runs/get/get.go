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
		Short: "Get information about a run",
		Example: heredoc.Doc(`
			airplane runs get <id>
			airplane runs get <id> -o yaml
			airplane runs get <id> -o json
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), c, args[0])
		},
	}
	return cmd
}

// Run runs the get command.
func run(ctx context.Context, c *cli.Config, id string) error {
	var client = c.Client

	resp, err := client.GetRun(ctx, id)
	if err != nil {
		return err
	}

	print.Run(resp.Run)
	return nil
}
