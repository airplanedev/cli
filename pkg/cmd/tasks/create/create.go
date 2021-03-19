package create

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// New returns a new create command.
func New(c *cli.Config) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		Long:  "Create a new task with a YAML configuration",
		Example: heredoc.Doc(`
			$ airplane tasks create -f task.yml
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Configuration file")
	cmd.MarkFlagRequired("file")

	return cmd
}

// Run runs the create command.
func run(ctx context.Context, c *cli.Config, file string) error {
	var client = c.Client
	var req api.CreateTaskRequest

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "read config %q", file)
	}

	if err := yaml.Unmarshal(buf, &req); err != nil {
		return errors.Wrapf(err, "unmarshal config")
	}

	res, err := client.CreateTask(ctx, req)
	if err != nil {
		return err
	}

	task, err := client.GetTask(ctx, res.Slug)
	if err != nil {
		return err
	}

	printTask(task, client.TaskURL(task.ID))
	return nil
}

// PrintTask prints the given task.
func printTask(t api.Task, url string) {
	fmt.Printf("\nCreated a task:\n\n")
	fmt.Printf("%s (%s)\n", t.Name, t.Slug)
	fmt.Printf("%s\n\n", t.Description)
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Arguments: %v\n\n", t.Arguments)
	fmt.Printf("To execute the task:\n")
	fmt.Printf("  airplane tasks execute %s [args]\n\n", t.Slug)
}
