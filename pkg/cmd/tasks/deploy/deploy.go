package deploy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	root *cli.Config
	file string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a task",
		Long:  "Deploy a task from a YAML-based task definition",
		Example: heredoc.Doc(`
			airplane tasks deploy -f my-task.yml
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a task definition file.")

	cli.Must(cmd.MarkFlagRequired("file"))

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var client = cfg.root.Client

	dir, err := taskdir.Open(cfg.file)
	if err != nil {
		return err
	}
	defer dir.Close()

	def, err := dir.ReadDefinition()
	if err != nil {
		return err
	}

	if def, err = def.Validate(); err != nil {
		return err
	}

	var taskID string
	task, err := client.GetTask(ctx, def.Slug)
	if err == nil {
		taskID = task.ID
	} else if aerr, ok := err.(api.Error); ok && aerr.Code == 404 {
		// A task with this slug does not exist, so we should create one.
		logger.Log("  Creating...")
		if res, err := client.CreateTask(ctx, api.CreateTaskRequest{
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
			return errors.Wrapf(err, "creating task %s", def.Slug)
		} else {
			taskID = res.TaskID
		}
	} else {
		return errors.Wrap(err, "getting task")
	}

	if def.Builder != "" {
		registry, err := client.GetRegistryToken(ctx)
		if err != nil {
			return errors.Wrap(err, "getting registry token")
		}

		var output io.Writer = ioutil.Discard
		if cfg.root.DebugMode {
			output = os.Stderr
		}

		b, err := build.New(build.Config{
			Root:    dir.Dir,
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

		logger.Log("  Building...")
		bo, err := b.Build(ctx, taskID, "latest")
		if err != nil {
			return errors.Wrap(err, "build")
		}

		logger.Log("  Updating...")
		if err := b.Push(ctx, bo.Tag); err != nil {
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

	logger.Log("  Done!")
	cmd := fmt.Sprintf("airplane tasks execute %s", def.Slug)
	if len(def.Parameters) > 0 {
		cmd += " -- [parameters]"
	}
	logger.Log(`
To execute %s:
- From the CLI: %s
- From the UI: %s`, def.Name, cmd, client.TaskURL(taskID))

	return nil
}
