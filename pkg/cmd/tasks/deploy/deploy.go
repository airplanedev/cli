package deploy

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	root    *cli.Config
	file    string
	builder string
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
			return run(cmd.Root().Context(), cfg)
		},
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a task definition file.")
	cmd.Flags().StringVar(&cfg.builder, "builder", string(build.BuilderKindRemote), "Where to build the task's Docker image. Accepts: [local, remote]")

	cli.Must(cmd.MarkFlagRequired("file"))

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var client = cfg.root.Client

	builder, err := build.ToBuilderKind(cfg.builder)
	if err != nil {
		return err
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

	if def, err = def.Validate(); err != nil {
		return err
	}

	if err := ensureConfigsExist(ctx, client, def); err != nil {
		return err
	}

	kind, kindOptions, err := def.GetKindAndOptions()
	if err != nil {
		return err
	}

	var image *string
	var command []string
	if def.Manual != nil {
		image = &def.Manual.Image
		command = def.Manual.Command
	}

	var taskID string
	task, err := client.GetTask(ctx, def.Slug)
	if err == nil {
		// This task already exists, so update it.
		logger.Log("Updating task...")
		_, err := client.UpdateTask(ctx, api.UpdateTaskRequest{
			// Only update kind/kindOptions since the remote build will
			// fetch this from the task.
			Kind:        kind,
			KindOptions: kindOptions,

			// The following fields are not updated until after the build finishes.
			//
			// TODO: move kind/kindOptions to builds so we don't need to do this.
			Slug:             task.Slug,
			Name:             task.Name,
			Description:      task.Description,
			Image:            task.Image,
			Command:          task.Command,
			Arguments:        task.Arguments,
			Parameters:       task.Parameters,
			Constraints:      task.Constraints,
			Env:              task.Env,
			ResourceRequests: task.ResourceRequests,
			Resources:        task.Resources,
			Repo:             task.Repo,
			Timeout:          task.Timeout,
		})
		if err != nil {
			return errors.Wrapf(err, "updating task %s", def.Slug)
		}

		taskID = task.ID
	} else if aerr, ok := err.(api.Error); ok && aerr.Code == 404 {
		// A task with this slug does not exist, so we should create one.
		logger.Log("Creating task...")
		res, err := client.CreateTask(ctx, api.CreateTaskRequest{
			Slug:             def.Slug,
			Name:             def.Name,
			Description:      def.Description,
			Image:            image,
			Command:          command,
			Arguments:        def.Arguments,
			Parameters:       def.Parameters,
			Constraints:      def.Constraints,
			Env:              def.Env,
			ResourceRequests: def.ResourceRequests,
			Resources:        def.Resources,
			Kind:             kind,
			KindOptions:      kindOptions,
			Repo:             def.Repo,
			Timeout:          def.Timeout,
		})
		if err != nil {
			return errors.Wrapf(err, "creating task %s", def.Slug)
		}

		taskID = res.TaskID
	} else {
		return errors.Wrap(err, "getting task")
	}

	if build.NeedsBuilding(kind) {
		resp, err := build.Run(ctx, build.Request{
			Builder: builder,
			Client:  client,
			Dir:     dir,
			Def:     def,
			TaskID:  taskID,
		})
		if err != nil {
			return err
		}

		_, err = client.UpdateTask(ctx, api.UpdateTaskRequest{
			Slug:             def.Slug,
			Name:             def.Name,
			Description:      def.Description,
			Image:            &resp.ImageURL,
			Command:          command,
			Arguments:        def.Arguments,
			Parameters:       def.Parameters,
			Constraints:      def.Constraints,
			Env:              def.Env,
			ResourceRequests: def.ResourceRequests,
			Resources:        def.Resources,
			Kind:             kind,
			KindOptions:      kindOptions,
			Repo:             def.Repo,
			Timeout:          def.Timeout,
		})
		if err != nil {
			return errors.Wrapf(err, "updating task %s", def.Slug)
		}
	}

	cmd := fmt.Sprintf("airplane execute %s", def.Slug)
	if len(def.Parameters) > 0 {
		cmd += " -- [parameters]"
	}
	logger.Log(`
To execute %s:
- From the CLI: %s
- From the UI: %s`, def.Name, cmd, client.TaskURL(taskID))

	return nil
}
