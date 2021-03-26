package utils

import "github.com/spf13/cobra"

// WithParentPersistentPreRunE runs the parent command's PersistentPreRunE before the current
// command's PersistentPreRunE. This prevents the default Cobra behavior of only running the
// final PersistentPreRunE.
func WithParentPersistentPreRunE(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for parent := cmd.Parent(); parent != nil; {
			// Find the first parent with a PersistentPreRunE, if any.
			if parent.PersistentPreRunE == nil {
				continue
			}

			if err := parent.PersistentPreRunE(parent, args); err != nil {
				return err
			}
			break
		}

		return f(cmd, args)
	}
}
