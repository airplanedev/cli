package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/airplanedev/cli/pkg/analytics"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
)

// DeployFromYaml deploys from a yaml file.
func deployFromYaml(ctx context.Context, cfg config) (err error) {
	client := cfg.client
	props := taskDeployedProps{
		from: "yaml",
	}
	start := time.Now()
	defer func() {
		analytics.Track(cfg.root, "Task Deployed", map[string]interface{}{
			"from":             props.from,
			"kind":             props.kind,
			"task_id":          props.taskID,
			"task_slug":        props.taskSlug,
			"task_name":        props.taskName,
			"build_id":         props.buildID,
			"errored":          err != nil,
			"duration_seconds": time.Since(start).Seconds(),
		})
	}()

	dir, err := taskdir.Open(cfg.file)
	if err != nil {
		return
	}
	defer dir.Close()

	def, err := dir.ReadDefinition()
	if err != nil {
		return
	}

	def, err = def.Validate()
	if err != nil {
		return
	}
	props.taskSlug = def.Slug

	err = ensureConfigsExist(ctx, client, def)
	if err != nil {
		return
	}

	kind, kindOptions, err := def.GetKindAndOptions()
	if err != nil {
		return
	}
	props.kind = kind

	// Remap resources from ref -> name to ref -> id.
	resp, err := client.ListResources(ctx)
	if err != nil {
		err = errors.Wrap(err, "fetching resources")
		return
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
			err = errors.Errorf("unknown resource: %s", name)
			return
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
		_, err = client.CreateTask(ctx, api.CreateTaskRequest{
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
			err = errors.Wrapf(err, "creating task %s", def.Slug)
			return
		}

		task, err = client.GetTask(ctx, def.Slug)
		if err != nil {
			err = errors.Wrap(err, "fetching created task")
			return
		}
	} else if err != nil {
		err = errors.Wrap(err, "getting task")
		return
	}
	props.taskID = task.ID
	props.taskName = task.Name

	if build.NeedsBuilding(kind) {
		resp, bErr := build.Run(ctx, build.Request{
			Local:  cfg.local,
			Client: client,
			Root:   dir.DefinitionRootPath(),
			Def:    def,
			TaskID: task.ID,
		})
		props.buildLocal = cfg.local
		props.buildID = resp.BuildID
		if bErr != nil {
			err = bErr
			return
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
		err = errors.Wrapf(err, "updating task %s", def.Slug)
		return
	}

	cmd := fmt.Sprintf("airplane exec %s", def.Slug)
	if len(def.Parameters) > 0 {
		cmd += " -- [parameters]"
	}

	logger.Suggest(
		"⚡ To execute the task from the CLI:",
		cmd,
	)

	logger.Suggest(
		"⚡ To execute the task from the UI:",
		client.TaskURL(def.Slug),
	)
}
