package push

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// New returns a new push command.
func New(c *cli.Config) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:     "push <slug>",
		Short:   "Push a task",
		Long:    "Push task with a YAML configuration",
		Example: "airplane push my-task -f task.yml",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, file, args[0])
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Configuration file")
	cmd.MarkFlagRequired("file")

	return cmd
}

// Run runs the create command.
func run(ctx context.Context, c *cli.Config, file, slug string) error {
	var client = c.Client
	var req api.UpdateTaskRequest

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "read config %s", file)
	}

	if err := yaml.Unmarshal(buf, &req); err != nil {
		return errors.Wrapf(err, "unmarshal config")
	}

	req.Slug = slug
	if err := client.UpdateTask(ctx, req); err != nil {
		return errors.Wrapf(err, "updating task %s", slug)
	}
	fmt.Println("  Updated", req.Slug)

	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return errors.Wrap(err, "getting task")
	}

	registry, err := client.GetRegistryToken(ctx)
	if err != nil {
		return errors.Wrap(err, "getting registry token")
	}

	root, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return err
	}

	b, err := build.New(build.Config{
		Root:    root,
		Builder: req.Builder,
		Args:    build.Args(req.BuilderConfig),
		Writer:  ioutil.Discard,
		Auth: &build.RegistryAuth{
			Token: registry.Token,
			Repo:  registry.Repo,
		},
	})
	if err != nil {
		return errors.Wrap(err, "new build")
	}

	fmt.Println("  Building...")
	img, err := b.Build(ctx, task.ID)
	if err != nil {
		return errors.Wrap(err, "build")
	}

	fmt.Println("  Pushing...")
	if err := b.Push(ctx, img.RepoTags[0]); err != nil {
		return errors.Wrap(err, "push")
	}

	fmt.Printf(`
  Updated the task %s, to execute it:

    airplane execute %s

`, req.Name, slug)
	return nil
}
