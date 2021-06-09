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
		return &unlinked{
			path: cfg.file,
		}
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
	var taskroot = filepath.Dir(abs)

	if root, err := r.Root(abs); err == nil {
		def.Node.Entrypoint = strings.TrimPrefix(abs, root)
		taskroot = root
	} else {
		def.Node.Entrypoint = filepath.Base(abs)
	}

	if wd, err := r.Workdir(abs); err == nil {
		def.Node.Workdir = strings.TrimPrefix(wd, taskroot)
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

	cmd := fmt.Sprintf("airplane execute %s", cfg.file)
	if len(def.Parameters) > 0 {
		cmd += " -- [parameters]"
	}
	logger.Log(`
To execute %s:
- From the CLI: %s
- From the UI: %s`, def.Name, cmd, client.TaskURL(task.Slug))
	return nil
}

// Unlinked explains an unlinked code error.
type unlinked struct {
	path string
}

// Error implementation.
func (u unlinked) Error() string {
	return fmt.Sprintf(
		"the file %s is not linked to a task",
		u.path,
	)
}

// ExplainError implementation.
func (u unlinked) ExplainError() string {
	return fmt.Sprintf(
		"You can link the file by running:\n  airplane init --slug <slug> %s",
		u.path,
	)
}
