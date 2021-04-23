package open

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Config is the open config.
type config struct {
	root *cli.Config
	slug string
	file string
}

// New returns a new open command.
func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Opens a task page in browser",
		Example: heredoc.Doc(`
			airplane tasks open <task_slug>
			airplane tasks open -f <task_definition.yml>
		`),
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				cfg.slug = args[0]
			}
			return run(cmd.Root().Context(), cfg)
		},
	}
	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a task definition file.")
	return cmd
}

// Run runs the open command.
func run(ctx context.Context, cfg config) error {
	var client = cfg.root.Client

	slug := cfg.slug
	if slug == "" {
		if cfg.file == "" {
			return errors.New("expected either a task slug or --file")
		}

		dir, err := taskdir.Open(cfg.file)
		if err != nil {
			return err
		}
		defer dir.Close()

		def, err := dir.ReadDefinition()
		if err != nil {
			return err
		}

		if def.Slug == "" {
			return errors.Errorf("no task slug found in task definition at %s", cfg.file)
		}

		slug = def.Slug
	}

	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return errors.Wrap(err, "get task")
	}

	taskURL := client.TaskURL(task.ID)
	logger.Log("Opening %s", taskURL)
	if !utils.Open(taskURL) {
		logger.Log("Could not open browser - try copying and pasting the above URL")
	}

	return nil
}
