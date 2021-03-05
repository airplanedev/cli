package root

import (
	"github.com/airplanedev/cli/commands/create"
	"github.com/airplanedev/cli/commands/list"
	"github.com/spf13/cobra"
)

// New returns a new root cobra command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airplane <command>",
		Short: "Airplane CLI",
	}

	cmd.AddCommand(create.New())
	cmd.AddCommand(list.New())

	return cmd
}
