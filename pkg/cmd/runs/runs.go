package runs

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/runs/get"
	"github.com/airplanedev/cli/pkg/cmd/runs/list"
	"github.com/spf13/cobra"
)

// New returns a new cobra command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Manage runs",
		Long:  "Manage runs",
		Example: heredoc.Doc(`
			airplane runs list --task my-task
			airplane runs get <id>
		`),
	}

	cmd.AddCommand(list.New(c))
	cmd.AddCommand(get.New(c))

	return cmd
}
