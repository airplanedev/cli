package info

import (
	"context"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	blue = color.New(color.FgHiBlue).SprintFunc()
	gray = color.New(color.FgHiBlack).SprintFunc()
)

func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get information about the currently logged in user / team",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c)
		},
	}
	return cmd
}

func run(ctx context.Context, c *cli.Config) error {
	var client = c.Client

	res, err := client.AuthInfo(ctx)
	if err != nil {
		return err
	}

	var userStr string
	if res.User == nil {
		userStr = gray("<no user>")
	} else {
		userStr = res.User.Email
	}
	logger.Log("  Signed in as %s", blue(userStr))
	logger.Log("  Using team %s (ID: %s)", blue(res.Team.Name), res.Team.ID)

	return nil
}
