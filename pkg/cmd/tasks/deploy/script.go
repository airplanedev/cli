package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/pkg/errors"
)

// DeployFromScript deploys from the given script.
func deployFromScript(ctx context.Context, cfg config) error {
	var client = cfg.client
	var ext = filepath.Ext(cfg.file)

	if ext == "" {
		return errors.New("cannot deploy a file without extension")
	}

	r, ok := runtime.Lookup(cfg.file)
	if !ok {
		return fmt.Errorf("cannot deploy a file with extension of %q", ext)
	}

	code, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		return fmt.Errorf("reading %s - %w", cfg.file, err)
	}

	slug, ok := runtime.Slug(code)
	if !ok {
		return runtime.ErrNotLinked{Path: cfg.file}
	}

	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return err
	}

	if task.Kind != r.Kind() {
		return fmt.Errorf("'%s' is a %s task. Expected a %s task.", task.Name, task.Kind, r.Kind())
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
		Slug:             def.Slug,
		Name:             def.Name,
		Description:      def.Description,
		Image:            &resp.ImageURL,
		Command:          []string{},
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
		return err
	}

	cmd := fmt.Sprintf("airplane exec %s", cfg.file)
	if len(def.Parameters) > 0 {
		cmd += " -- [parameters]"
	}

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
	default:
		panic(fmt.Sprintf("setEntrypoint received unexpected kind %q", kind))
	}
}
