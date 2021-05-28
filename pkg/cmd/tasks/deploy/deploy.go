package deploy

import (
	"context"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

type config struct {
	client *api.Client
	file   string
	local  bool
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{client: c.Client}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a task",
		Long:  "Deploy a task from a YAML-based task definition",
		Example: heredoc.Doc(`
			airplane tasks deploy my-task.yml
			airplane tasks deploy task.ts
			airplane tasks deploy task.js
			airplane tasks deploy --local task.js
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.file = args[0]
			return run(cmd.Root().Context(), cfg)
		},
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
	}

	cmd.Flags().BoolVarP(&cfg.local, "local", "L", false, "Add to build the docker image locally.")

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var ext = filepath.Ext(cfg.file)

	if ext == ".yml" || ext == ".yaml" {
		return deployFromYaml(ctx, cfg)
	}

	return deployFromScript(ctx, cfg)
}
