package execute

import (
	"context"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new execute cobra command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "Execute a task",
		Long:    "Execute a task by its slug with the provided arguments.",
		Example: "airplane execute echo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c, args)
		},
	}
	return cmd
}

// Run runs the execute command.
func run(ctx context.Context, c *cli.Config, args []string) error {
	var client = c.Client

	task, err := client.GetTask(ctx, args[0])
	if err != nil {
		return errors.Wrap(err, "get task")
	}

	fmt.Println()
	fmt.Println("  Execute " + task.Name + ":")
	fmt.Println()

	var (
		req = api.RunTaskRequest{
			TaskID:     task.ID,
			Parameters: make(api.Values),
		}
	)

	if err := survey.Ask(questions(task.Parameters), &req.Parameters); err != nil {
		return err
	}

	run, err := client.RunTask(ctx, req)
	if err != nil {
		return errors.Wrap(err, "run task")
	}

	fmt.Printf("  Queued: %s\n", client.RunURL(run.RunID))

	var resp api.GetRunResponse

	for {
		resp, err = client.GetRun(ctx, run.RunID)
		if err != nil {
			return errors.Wrap(err, "get run")
		}

		var done bool

		switch resp.Run.Status {
		case api.RunSucceeded,
			api.RunCancelled,
			api.RunFailed:
			done = true
		}

		if done {
			break
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Printf("  Done: %s\n", resp.Run.Status)
	fmt.Println()
	return nil
}

// Questions accepts a slice of parameters and returns survey questions.
func questions(params api.Parameters) []*survey.Question {
	var ret = make([]*survey.Question, 0, len(params))

	for _, p := range params {
		ret = append(ret, &survey.Question{
			Name: p.Slug,
			Prompt: &survey.Input{
				Message: p.Name,
				Help:    p.Desc,
			},
		})
	}

	return ret
}
