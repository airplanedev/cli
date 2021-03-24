package deploy

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
	"github.com/airplanedev/cli/pkg/taskdef"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	cli   *cli.Config
	debug bool
	file  string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{cli: c}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a task",
		Long:  "Deploy a task from a YAML-based task definition",
		Example: heredoc.Doc(`
			$ airplane tasks deploy -f my-task.yml
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().BoolVar(&cfg.debug, "debug", false, "Print extra debug logging while building images.")
	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a task definition file.")

	cli.Must(cmd.MarkFlagRequired("file"))

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var client = cfg.cli.Client

	def, err := taskdef.Read(cfg.file)
	if err != nil {
		return err
	}

	var taskID string
	// If a slug is not set in the task definition, then we'll create a brand new
	// task for this definition.
	if def.Slug == "" {
		fmt.Println("  Creating...")
		if res, err := client.CreateTask(ctx, api.CreateTaskRequest{
			Name:           def.Name,
			Description:    def.Description,
			Image:          def.Image,
			Command:        def.Command,
			Arguments:      def.Arguments,
			Parameters:     def.Parameters,
			Constraints:    def.Constraints,
			Env:            def.Env,
			ResourceLimits: def.ResourceLimits,
			Builder:        def.Builder,
			BuilderConfig:  def.BuilderConfig,
			Repo:           def.Repo,
			Timeout:        def.Timeout,
		}); err != nil {
			return errors.Wrapf(err, "updating task %s", def.Slug)
		} else {
			taskID = res.TaskID
			def.Slug = res.Slug
		}

		// Insert the new slug into the task definitions so that future
		// calls to deploy will reference the task we just created.
		if err := taskdef.Write(cfg.file, def); err != nil {
			return errors.Wrap(err, "updating task definition with slug")
		}
	} else {
		task, err := client.GetTask(ctx, def.Slug)
		if err != nil {
			return errors.Wrap(err, "getting task")
		}
		taskID = task.ID
	}

	if def.Builder != "" {
		registry, err := client.GetRegistryToken(ctx)
		if err != nil {
			return errors.Wrap(err, "getting registry token")
		}

		root, err := filepath.Abs(filepath.Dir(cfg.file))
		if err != nil {
			return errors.Wrap(err, "getting root directory")
		}

		var output io.Writer = ioutil.Discard
		if cfg.debug {
			output = os.Stderr
		}

		b, err := build.New(build.Config{
			Root:    root,
			Builder: def.Builder,
			Args:    build.Args(def.BuilderConfig),
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
		img, err := b.Build(ctx, taskID, "latest")
		if err != nil {
			return errors.Wrap(err, "build")
		}

		fmt.Println("  Updating...")
		if err := b.Push(ctx, img.RepoTags[0]); err != nil {
			return errors.Wrap(err, "push")
		}
	}

	if err := client.UpdateTask(ctx, api.UpdateTaskRequest{
		Slug:           def.Slug,
		Name:           def.Name,
		Description:    def.Description,
		Image:          def.Image,
		Command:        def.Command,
		Arguments:      def.Arguments,
		Parameters:     def.Parameters,
		Constraints:    def.Constraints,
		Env:            def.Env,
		ResourceLimits: def.ResourceLimits,
		Builder:        def.Builder,
		BuilderConfig:  def.BuilderConfig,
		Repo:           def.Repo,
		Timeout:        def.Timeout,
	}); err != nil {
		return errors.Wrapf(err, "updating task %s", def.Slug)
	}

	fmt.Println("  Done!")
	fmt.Printf(`
To execute %s:
- From the UI: %s
- From the CLI: airplane tasks execute %s -- [parameters]
`, def.Name, client.TaskURL(taskID), def.Slug)

	return nil
}
