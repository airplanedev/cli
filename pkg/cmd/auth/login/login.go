package login

import (
	"context"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/token"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

// New returns a new login command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Airplane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cmd, c)
		},
	}
	return cmd
}

// Run runs the login command.
func run(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
	if err := EnsureLoggedIn(ctx, cmd, c); err != nil {
		return err
	}

	cmd.Printf("You're all set!\n\nTo see what tasks you can run, try:\n    airplane tasks list\n")
	return nil
}

func EnsureLoggedIn(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
	if c.Client.Token != "" {
		return nil
	}

	srv, err := token.NewServer(ctx)
	if err != nil {
		return err
	}
	defer srv.Close()

	url := c.Client.LoginURL(srv.URL())
	if ok := utils.Open(url); !ok {
		cmd.Printf("Visit %s to complete logging in\n", url)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()

	case token := <-srv.Token():
		c.Client.Token = token
		cfg, err := conf.ReadDefault()
		if err != nil {
			return err
		}
		cfg.Token = token
		if err := conf.WriteDefault(cfg); err != nil {
			return err
		}
	}

	return nil
}
