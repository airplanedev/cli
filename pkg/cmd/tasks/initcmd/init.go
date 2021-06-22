// initcmd defines the implementation of the `airplane tasks init` command.
//
// Even though the command is called "init", we can't name the package "init"
// since that conflicts with the Go init function.
package initcmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/runtime"
	_ "github.com/airplanedev/cli/pkg/runtime/javascript"
	_ "github.com/airplanedev/cli/pkg/runtime/python"
	_ "github.com/airplanedev/cli/pkg/runtime/typescript"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/spf13/cobra"
)

type config struct {
	client *api.Client
	file   string
	slug   string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{client: c.Client}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a task definition",
		Example: heredoc.Doc(`
			$ airplane tasks init
			$ airplane tasks init --slug task-slug ./my/task.js
			$ airplane tasks init --slug task-slug ./my/task.ts
		`),
		Args: cobra.ExactArgs(1),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.file = args[0]
			return run(cmd.Root().Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.slug, "slug", "", "Slug of an existing task to generate from.")
	cmd.MarkFlagRequired("slug")

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var ext = filepath.Ext(cfg.file)
	var client = cfg.client

	if ext == "" {
		return fmt.Errorf("expected <path> %q to have a file extension", cfg.file)
	}

	r, ok := runtime.Lookup(cfg.file)
	if !ok {
		return fmt.Errorf("unable to deploy task with %q file extension", ext)
	}

	task, err := client.GetTask(ctx, cfg.slug)
	if err != nil {
		return err
	}

	if task.Kind != r.Kind() {
		return fmt.Errorf("cannot link %q to a %s task", cfg.file, r.Kind())
	}

	if fsx.Exists(cfg.file) {
		buf, err := ioutil.ReadFile(cfg.file)
		if err != nil {
			return err
		}

		if slug, ok := runtime.Slug(buf); ok && slug == task.Slug {
			logger.Log("%s is already linked to %s", cfg.file, cfg.slug)
			suggestDeploy(cfg.file)
			return nil
		}

		patch, err := patch(cfg.slug, cfg.file)
		if err != nil {
			return err
		}

		if !patch {
			logger.Log("You canceled linking %s to %s", cfg.file, cfg.slug)
			return nil
		}

		code := []byte(runtime.Comment(r, task))
		code = append(code, '\n', '\n')
		code = append(code, buf...)

		if err := ioutil.WriteFile(cfg.file, code, 0644); err != nil {
			return err
		}

		logger.Log("Linked %s to %s", cfg.file, cfg.slug)
		suggestDeploy(cfg.file)
		return nil
	}

	code, err := r.Generate(task)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfg.file), 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(cfg.file, code, 0644); err != nil {
		return err
	}

	logger.Log("Initialized a task at %s", cfg.file)
	suggestDeploy(cfg.file)
	return nil
}

// SuggestDeploy suggests a deploy to the user.
func suggestDeploy(file string) {
	logger.Log("You can deploy this task with:")
	logger.Log("  airplane deploy %s", file)
}

// Patch asks the user if he would like to patch a file
// and add the airplane special comment.
func patch(slug, file string) (ok bool, err error) {
	err = survey.AskOne(
		&survey.Confirm{
			Message: fmt.Sprintf("Would you like to link %s to %s?", file, slug),
			Help:    "Linking this file will add a special airplane comment.",
			Default: false,
		},
		&ok,
	)
	return
}
