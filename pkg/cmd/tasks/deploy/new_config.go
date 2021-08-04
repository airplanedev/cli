package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/pkg/errors"
)

func deployFromNewConfig(ctx context.Context, cfg config) error {
	client := cfg.client
	ext := filepath.Ext(cfg.file)

	// if file ends in *.task.yaml|yml|json, expand it in buildConfig
	// buildConfig manages entrypoint, slug, and optional kindOptions

	r, ok := runtime.Lookup(cfg.file)
	if !ok {
		return errors.Errorf("cannot deploy a file with extension of %q", ext)
	}

	code, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		return errors.Wrapf(err, "reading %s", cfg.file)
	}

	slug, ok := runtime.Slug(code)
	if !ok {
		return runtime.ErrNotLinked{Path: cfg.file}
	}
	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return err
	}

	def, err := definitions.NewDefinitionFromTask(task)
	if err != nil {
		return err
	}
	kind, kindOptions, err := def.GetKindAndOptions()
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(cfg.file)
	if err != nil {
		return err
	}
	taskroot, err := r.Root(abs)
	if err != nil {
		return err
	}
	entrypoint, err := filepath.Rel(taskroot, abs)
	if err != nil {
		return err
	}
	setEntrypoint(&def, entrypoint)

	// Figure out BuildConfig struct here
	// Pass it to build.Run
	resp, err := build.Run(ctx, build.Request{
		Local:   cfg.local,
		Client:  client,
		TaskID:  task.ID,
		Root:    taskroot,
		Def:     def,
		TaskEnv: def.Env,
		Shim:    true,
	})
	if err != nil {
		return err
	}

	_, err = client.UpdateTask(ctx, api.UpdateTaskRequest{
		Slug:                       def.Slug,
		Name:                       def.Name,
		Description:                def.Description,
		Image:                      &resp.ImageURL,
		Command:                    []string{},
		Arguments:                  def.Arguments,
		Parameters:                 def.Parameters,
		Constraints:                def.Constraints,
		Env:                        def.Env,
		ResourceRequests:           def.ResourceRequests,
		Resources:                  def.Resources,
		Kind:                       kind,
		KindOptions:                kindOptions,
		Repo:                       def.Repo,
		RequireExplicitPermissions: task.RequireExplicitPermissions,
		Permissions:                task.Permissions,
		Timeout:                    def.Timeout,
	})
	if err != nil {
		return err
	}

	// Leave off `-- [parameters]` for simplicity - user will get prompted.
	cmd := fmt.Sprintf("airplane exec %s", cfg.file)
	logger.Suggest(
		"⚡ To execute the task from the CLI:",
		cmd,
	)

	logger.Suggest(
		"⚡ To execute the task from the UI:",
		client.TaskURL(task.Slug),
	)
	return nil
}
