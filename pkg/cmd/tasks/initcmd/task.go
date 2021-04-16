package initcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
)

func initFromTask(ctx context.Context, cfg config) error {
	client := cfg.root.Client

	var task api.Task
	var err error
	if cfg.from != "" {
		if task, err = client.GetTask(ctx, cfg.from); err != nil {
			return errors.Wrap(err, "getting task")
		}
	} else {
		if task, err = pickTask(ctx, client); err != nil {
			return err
		}
	}

	file := cfg.file
	if file == "" {
		file = "airplane.yml"
	}
	dir, err := taskdir.Open(file)
	if err != nil {
		return errors.Wrap(err, "opening task directory")
	}
	defer dir.Close()

	r, err := client.GetUniqueSlug(ctx, task.Name, task.Slug)
	if err != nil {
		return errors.Wrap(err, "getting unique slug")
	}

	if err := dir.WriteDefinition(taskdir.Definition{
		Slug:           r.Slug,
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

	logger.Log(`
An Airplane task definition for '%s' has been created!

To deploy it to Airplane, run:
  airplane deploy -f %s`, task.Name, file)

	return nil
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
