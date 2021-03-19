package push

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Config represents push configuration.
type config struct {
	cli     *cli.Config
	slug    string
	debug   bool
	version string
	file    string
}

// New returns a new push command.
func New(c *cli.Config) *cobra.Command {
	var cfg = config{cli: c}

	cmd := &cobra.Command{
		Use:   "push <slug>",
		Short: "Push a task",
		Long:  "Push task with a YAML configuration",
		Example: heredoc.Doc(`
			$ airplane tasks push my-task -f my-task.yml
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.slug = args[0]
			return run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.version, "tag", "t", "latest", "The version of the image.")
	cmd.Flags().BoolVarP(&cfg.debug, "debug", "", false, "Debug docker builds.")
	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Configuration file.")

	cmd.MarkFlagRequired("file")

	return cmd
}

// Run runs the create command.
func run(ctx context.Context, cfg config) error {
	var client = cfg.cli.Client
	var req api.UpdateTaskRequest

	task, err := client.GetTask(ctx, cfg.slug)
	if err != nil {
		return errors.Wrap(err, "get task")
	}

	buf, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		return errors.Wrapf(err, "read config %s", cfg.file)
	}

	if err := yaml.Unmarshal(buf, &req); err != nil {
		return errors.Wrapf(err, "unmarshal config")
	}

	if req.Builder != "" {
		registry, err := client.GetRegistryToken(ctx)
		if err != nil {
			return errors.Wrap(err, "getting registry token")
		}

		root, err := filepath.Abs(filepath.Dir(cfg.file))
		if err != nil {
			return err
		}

		var output io.Writer = ioutil.Discard
		if cfg.debug {
			output = os.Stderr
		}

		b, err := build.New(build.Config{
			Root:    root,
			Builder: req.Builder,
			Args:    build.Args(req.BuilderConfig),
			Writer:  output,
			Auth: &build.RegistryAuth{
				Token: registry.Token,
				Repo:  registry.Repo,
			},
		})
		if err != nil {
			return errors.Wrap(err, "new build")
		}

		fmt.Println("  Building...")
		img, err := b.Build(ctx, task.ID, cfg.version)
		if err != nil {
			return errors.Wrap(err, "build")
		}

		fmt.Println("  Pushing...")
		if err := b.Push(ctx, img.RepoTags[0]); err != nil {
			return errors.Wrap(err, "push")
		}
	}

	req.Slug = cfg.slug
	if err := client.UpdateTask(ctx, req); err != nil {
		return errors.Wrapf(err, "updating task %s", cfg.slug)
	}
	fmt.Println("  Updated", req.Slug)

	fmt.Printf(`
  Updated the task %s, to execute it:

    airplane tasks execute %s
`, req.Name, cfg.slug)
	return nil
}
