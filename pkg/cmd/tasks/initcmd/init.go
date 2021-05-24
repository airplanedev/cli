// initcmd defines the implementation of the `airplane tasks init` command.
//
// Even though the command is called "init", we can't name the package "init"
// since that conflicts with the Go init function.
package initcmd

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/scaffold"
	_ "github.com/airplanedev/cli/pkg/scaffold/typescript"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

type config struct {
	client *api.Client
	file   string
	slug   string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{client: c.Client}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a task definition",
		Example: heredoc.Doc(`
			$ airplane tasks init
			$ airplane tasks init --slug task-slug ./my/task.js
			$ airplane tasks init --slug task-slug ./my/task.ts
		`),
		Args: cobra.ExactArgs(1),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.file = args[0]
			return run(cmd.Root().Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.slug, "slug", "", "Slug of an existing task to generate from.")

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var client = cfg.client

	task, err := client.GetTask(ctx, cfg.slug)
	if err != nil {
		return err
	}

	if err := scaffold.Generate(cfg.file, task); err != nil {
		return err
	}

	return nil
}
