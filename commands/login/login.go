package login

import (
	"context"

	"github.com/spf13/cobra"
)

// New returns a new login command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Airplane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}
	return cmd
}

// Run runs the login command.
func run(ctx context.Context) error {
	return nil
}
