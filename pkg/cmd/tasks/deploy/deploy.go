package deploy

import (
	"context"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
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
		Long:  "Deploy code from a local directory to Airplane.",
		Example: heredoc.Doc(`
			airplane tasks deploy ./task.ts
			airplane tasks deploy --local ./task.js
			airplane tasks deploy ./my-task.yml
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.file != "" {
				// A file was provided with the -f flag. This is deprecated.
				logger.Warning(`The --file/-f flag is deprecated and will be removed in a future release. File paths should be passed as a positional argument instead: airplane deploy %s`, cfg.file)
			} else if len(args) > 0 {
				cfg.file = args[0]
			} else {
				return errors.New("expected 1 argument: airplane deploy ./path/to/file")
			}
			return run(cmd.Root().Context(), cfg)
		},
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
	}

	cmd.Flags().BoolVarP(&cfg.local, "local", "L", false, "use a local Docker daemon (instead of an Airplane-hosted builder)")
	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "File to deploy (.yaml, .yml, .js, .ts)")
	cli.Must(cmd.Flags().MarkHidden("file")) // --file is deprecated

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var ext = filepath.Ext(cfg.file)

	if ext == ".yml" || ext == ".yaml" {
		return deployFromYaml(ctx, cfg)
	}

	return deployFromScript(ctx, cfg)
}
