package deploy

import (
	"context"
	"fmt"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
)

// DeployFromYaml deploys from a yaml file.
func deployFromYaml(ctx context.Context, cfg config) error {
	var client = cfg.client

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

	// Remap resources from ref -> name to ref -> id.
	resp, err := client.ListResources(ctx)
	if err != nil {
		return errors.Wrap(err, "fetching resources")
	}
	resourcesByName := map[string]api.Resource{}
	for _, resource := range resp.Resources {
		resourcesByName[resource.Name] = resource
	}
	resources := map[string]string{}
	for ref, name := range def.Resources {
		if res, ok := resourcesByName[name]; ok {
			resources[ref] = res.ID
		} else {
			return errors.Errorf("unknown resource: %s", name)
		}
	}

	var image *string
	var command []string
	if def.Image != nil {
		image = &def.Image.Image
		command = def.Image.Command
	}

	task, err := client.GetTask(ctx, def.Slug)
	if _, ok := err.(*api.TaskMissingError); ok {
		// A task with this slug does not exist, so we should create one.
		logger.Log("Creating task...")
		_, err := client.CreateTask(ctx, api.CreateTaskRequest{
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

		task, err = client.GetTask(ctx, def.Slug)
		if err != nil {
			return errors.Wrap(err, "fetching created task")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting task")
	}

	if build.NeedsBuilding(kind) {
		// Before performing a remote build, we must first update kind/kindOptions
		// since the remote build relies on pulling those from the tasks table (for now).
		_, err := client.UpdateTask(ctx, api.UpdateTaskRequest{
			Kind:        kind,
			KindOptions: kindOptions,

			// The following fields are not updated until after the build finishes.
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

		resp, err := build.Run(ctx, build.Request{
			Local:  cfg.local,
			Client: client,
			Root:   dir.DefinitionRootPath(),
			Def:    def,
			TaskID: task.ID,
		})
		if err != nil {
			return err
		}
		image = &resp.ImageURL
	}

	_, err = client.UpdateTask(ctx, api.UpdateTaskRequest{
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
		Resources:        resources,
		Kind:             kind,
		KindOptions:      kindOptions,
		Repo:             def.Repo,
		Timeout:          def.Timeout,
	})
	if err != nil {
		return errors.Wrapf(err, "updating task %s", def.Slug)
	}

	cmd := fmt.Sprintf("airplane execute %s", def.Slug)
	if len(def.Parameters) > 0 {
		cmd += " -- [parameters]"
	}
	logger.Log(`
To execute %s:
- From the CLI: %s
- From the UI: %s`, def.Name, cmd, client.TaskURL(def.Slug))

	return nil
}
