package create

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Config represents create config.
type config struct {
	file string
}

// New returns a new create command.
func New() *cobra.Command {
	var cfg config

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a task",
		Long:    "Create a new task with a YAML configuration",
		Example: "airplane create -f task.yml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Configuration file")
	cmd.MarkFlagRequired("file")

	return cmd
}

// Run runs the create command.
func run(ctx context.Context, cfg config) error {
	var client api.Client
	var req api.CreateTaskRequest

	buf, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		return errors.Wrapf(err, "read config %q", cfg.file)
	}

	if err := yaml.Unmarshal(buf, &req); err != nil {
		return errors.Wrapf(err, "unmarshal config")
	}

	if res, err := client.CreateTask(ctx, req); err != nil {
		return errors.Wrapf(err, "create task")
	} else {
		fmt.Println("created task", res.Slug)
	}

	return nil
}
