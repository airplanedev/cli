// initcmd defines the implementation of the `airplane tasks init` command.
//
// Even though the command is called "init", we can't name the package "init"
// since that conflicts with the Go init function.
package initcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	cli  *cli.Config
	file string
	from string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{cli: c}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a task definition",
		Example: heredoc.Doc(`
			$ airplane tasks init
			$ airplane tasks init -f ./airplane.yml
			$ airplane tasks init --from hello_world
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cmd, cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "airplane.yml", "Path to a file to store task definition")
	cmd.Flags().StringVar(&cfg.from, "from", "", "Slug of an existing task to generate from")

	return cmd
}

func run(ctx context.Context, cmd *cobra.Command, cfg config) error {
	client := cfg.cli.Client

	var kind initKind
	var err error
	// If --from is provided, we already know the user wants to create
	// from an existing task, so we don't need to prompt the user here.
	if cfg.from == "" {
		cmd.Printf("Airplane is the code-first solution for engineers building internal tools.\n\n")
		cmd.Printf("This command will configure a task definition which specifies how to deploy your task to Airplane.\n\n")

		if kind, err = pickInitKind(); err != nil {
			return err
		}
	}

	var taskName string
	switch kind {
	case initKindSample:
		// TODO
	case initKindScratch:
		// TODO
	case initKindExisting:
		var task api.Task
		if cfg.from != "" {
			if task, err = client.GetTask(ctx, cfg.from); err != nil {
				return errors.Wrap(err, "getting task")
			}
		} else {
			if task, err = pickTask(ctx, client); err != nil {
				return err
			}
		}

		dir, err := taskdir.Open(cfg.file)
		if err != nil {
			return errors.Wrap(err, "opening task directory")
		}
		defer dir.Close()

		if err := dir.WriteDefinition(taskdir.Definition{
			Slug:           task.Slug,
			Name:           task.Name,
			Description:    task.Description,
			Image:          task.Image,
			Command:        task.Command,
			Arguments:      task.Arguments,
			Parameters:     task.Parameters,
			Constraints:    task.Constraints,
			Env:            task.Env,
			ResourceLimits: task.ResourceLimits,
			Builder:        task.Builder,
			BuilderConfig:  task.BuilderConfig,
			Repo:           task.Repo,
			Timeout:        task.Timeout,
		}); err != nil {
			return errors.Wrap(err, "writing task definition")
		}

		taskName = task.Name
	default:
		return errors.Errorf("Unexpected unknown initKind choice: %s", kind)
	}

	cmd.Printf("\nAn Airplane task definition for '%s' has been created in %s!\n\nTo deploy it to Airplane, run:\n  airplane tasks deploy -f %s", taskName, cfg.file, cfg.file)

	return nil
}

type initKind string

var (
	initKindSample   initKind = "Create from an Airplane-provided sample"
	initKindScratch  initKind = "Create from scratch"
	initKindExisting initKind = "Create from an existing Airplane task"
)

func pickInitKind() (initKind, error) {
	var kind string
	if err := survey.AskOne(
		&survey.Select{
			// TODO: idk what do we say here
			Message: "How do you want to get started?",
			// TODO: upstream the ability to disable Survey's search filter
			Options: []string{
				string(initKindSample),
				string(initKindScratch),
				string(initKindExisting),
			},
			Default: string(initKindSample),
			// Help:    "todo",
		},
		&kind,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return initKind(""), errors.Wrap(err, "selecting kind of init")
	}

	return initKind(kind), nil
}

func pickTask(ctx context.Context, client *api.Client) (api.Task, error) {
	tasks, err := client.ListTasks(ctx)
	if err != nil {
		return api.Task{}, err
	}

	options := []string{}
	optionsToTask := map[string]*api.Task{}
	for i, task := range tasks.Tasks {
		option := fmt.Sprintf("%s (%s)", task.Name, task.Slug)
		options = append(options, option)
		optionsToTask[option] = &tasks.Tasks[i]
	}

	var selected string
	if err := survey.AskOne(
		&survey.Select{
			Message: "Choose a task:",
			Options: options,
		},
		&selected,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return api.Task{}, errors.Wrap(err, "selecting task to init from")
	}

	task, ok := optionsToTask[selected]
	if !ok || task == nil {
		return api.Task{}, errors.Wrap(err, "unexpected task selected")
	}

	return *task, nil
}

/**

Intro:

Airplane allows you to do X Y and Z.

This command will configure a *task definition* which is used to deploy tasks to Airplane.

Choose [create from existing task; create from samples; create new (?)]

Do you want to create a task definition from an existing task? y/N

[existing]: login, if not already; pick a task from list; dump in file

[sample]: pick a language; pick an example

[scratch]: pick a language; pick a name, desc (arguments??)

Hmm should we have a reference somewhere?s

*/
