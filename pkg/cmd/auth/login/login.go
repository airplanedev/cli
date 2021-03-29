package login

import (
	"context"
	"errors"

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
	if !isLoggedIn(c) {
		if err := login(ctx, cmd, c); err != nil {
			return err
		}
	}

	cmd.Printf("You're all set!\n\nTo see what tasks you can run, try:\n    airplane tasks list\n")
	return nil
}

var (
	ErrLoggedOut = errors.New("You are not logged in. To login, run:\n    airplane auth login")
)

func EnsureLoggedIn(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
	if isLoggedIn(c) {
		return nil
	}

	if !utils.CanPrompt() {
		return ErrLoggedOut
	}

	if ok, err := utils.Confirm("You are not logged in. Do you want to login now?"); err != nil {
		return err
	} else if !ok {
		return ErrLoggedOut
	}

	cmd.Printf("\n  Logging in...\n\n")

	if err := login(ctx, cmd, c); err != nil {
		return err
	}

	return nil
}

func isLoggedIn(c *cli.Config) bool {
	return c.Client.Token != ""
}

func login(ctx context.Context, cmd *cobra.Command, c *cli.Config) error {
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
