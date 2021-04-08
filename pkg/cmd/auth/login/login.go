package login

import (
	"context"
	"errors"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/logger"
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
			return run(cmd.Root().Context(), c)
		},
	}
	return cmd
}

// Run runs the login command.
func run(ctx context.Context, c *cli.Config) error {
	if !isLoggedIn(c) {
		if err := login(ctx, c); err != nil {
			return err
		}
	}

	logger.Log("You're all set!\n\nTo see what tasks you can run, try:\n    airplane tasks list")
	return nil
}

var (
	ErrLoggedOut = errors.New("You are not logged in. To login, run:\n    airplane login")
)

func EnsureLoggedIn(ctx context.Context, c *cli.Config) error {
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

	logger.Log("\n  Logging in...\n")

	if err := login(ctx, c); err != nil {
		return err
	}

	return nil
}

func isLoggedIn(c *cli.Config) bool {
	return c.Client.Token != ""
}

func login(ctx context.Context, c *cli.Config) error {
	srv, err := token.NewServer(ctx)
	if err != nil {
		return err
	}
	defer srv.Close()

	url := c.Client.LoginURL(srv.URL())
	if ok := utils.Open(url); !ok {
		logger.Log("Visit %s to complete logging in", url)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()

	case token := <-srv.Token():
		c.Client.Token = token
		cfg, err := conf.ReadDefault()
		if err != nil && !errors.Is(err, conf.ErrMissing) {
			return err
		}
		if cfg.Tokens == nil {
			cfg.Tokens = map[string]string{}
		}
		cfg.Tokens[c.Client.Host] = token
		if err := conf.WriteDefault(cfg); err != nil {
			return err
		}
	}

	return nil
}
