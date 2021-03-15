package push

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/airplanedev/cli/pkg/api"
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

	fmt.Printf(`
  Updated the task %s, to execute it:

    airplane execute %s

`, req.Name, slug)
	return nil
}
