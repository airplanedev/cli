package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/airplanedev/cli/pkg/analytics"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/pkg/errors"
)

// DeployFromScript deploys from the given script.
func deployFromScript(ctx context.Context, cfg config) (rErr error) {
	client := cfg.client
	tp := taskDeployedProps{
		from: "script",
	}
	start := time.Now()
	defer func() {
		analytics.Track(cfg.root, "Task Deployed", map[string]interface{}{
			"from":             tp.from,
			"kind":             tp.kind,
			"task_id":          tp.taskID,
			"task_slug":        tp.taskSlug,
			"task_name":        tp.taskName,
			"build_id":         tp.buildID,
			"errored":          rErr != nil,
			"duration_seconds": time.Since(start).Seconds(),
		})
	}()

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
	tp.kind = task.Kind
	tp.taskID = task.ID
	tp.taskSlug = task.Slug
	tp.taskName = task.Name

	r, err := runtime.Lookup(task.Kind, cfg.file)
	if err != nil {
		return errors.Wrapf(err, "cannot determine how to deploy %q - check your CLI is up to date", cfg.file)
	}

	def, err := definitions.NewDefinitionFromTask(task)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(cfg.file)
	if err != nil {
		return err
	}

	// Detect the root of the task, if found ensure
	// that the entrypoint and the root are included
	// in the build.
	taskroot, err := r.Root(abs)
	if err != nil {
		return err
	}
	entrypoint, err := filepath.Rel(taskroot, abs)
	if err != nil {
		return err
	}
	setEntrypoint(&def, entrypoint)

	// TODO(amir): move to `d.SetWorkdir()`.
	if def.Node != nil {
		if wd, err := r.Workdir(abs); err == nil {
			def.Node.Workdir = strings.TrimPrefix(wd, taskroot)
		}
	}

	kind, kindOptions, err := def.GetKindAndOptions()
	if err != nil {
		return err
	}

	tp.buildLocal = cfg.local
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
	tp.buildID = resp.BuildID

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

// SetEntrypoint sets the entrypoint on d.
//
// TODO(amir): move this to `def.SetEntrypoint()` or whatever.
func setEntrypoint(d *definitions.Definition, ep string) {
	switch kind, _, _ := d.GetKindAndOptions(); kind {
	case api.TaskKindNode:
		d.Node.Entrypoint = ep
	case api.TaskKindPython:
		d.Python.Entrypoint = ep
	case api.TaskKindShell:
		d.Shell.Entrypoint = ep
	default:
		panic(fmt.Sprintf("setEntrypoint received unexpected kind %q", kind))
	}
}
