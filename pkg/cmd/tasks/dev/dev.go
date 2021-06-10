package dev

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/fs"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/params"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	root *cli.Config
	file string
	args []string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}

	cmd := &cobra.Command{
		Use:     "dev ./path/to/file",
		Short:   "Locally execute a task",
		Aliases: []string{"exec"},
		Long:    "Locally executes a task, optionally with specific parameters.",
		Example: heredoc.Doc(`
			airplane dev ./task.js [-- <parameters...>]
			airplane dev ./task.ts [-- <parameters...>]
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New(`expected a file: airplane dev ./path/to/file`)
			}
			cfg.file = args[0]
			cfg.args = args[1:]

			return run(cmd.Root().Context(), cfg)
		},
	}

	return cmd
}

// Run runs the execute command.
func run(ctx context.Context, cfg config) error {
	if !fs.Exists(cfg.file) {
		return errors.Errorf("Unable to open file: %s", cfg.file)
	}

	slug, err := slugFromScript(cfg.file)
	if err != nil {
		return err
	}

	task, err := cfg.root.Client.GetTask(ctx, slug)
	if err != nil {
		return errors.Wrap(err, "getting task")
	}

	r, ok := runtime.Lookup(cfg.file)
	if !ok {
		return errors.Errorf("Unsupported file type: %s", filepath.Base(cfg.file))
	}

	paramValues, err := params.CLI(cfg.args, cfg.root.Client, task)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	} else if err != nil {
		return err
	}

	logger.Log("Locally executing %s: %s", logger.Bold(task.Name), logger.Gray(cfg.root.Client.TaskURL(task.Slug)))

	cmds, err := r.PrepareRun(ctx, paramValues)
	if err != nil {
		return err
	}

	logger.Log("")

	cmd := exec.CommandContext(ctx, cmds[0], cmds[1:]...)
	// TODO: output parsing
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// slugFromScript attempts to extract a slug from a script.
func slugFromScript(file string) (string, error) {
	code, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s - %w", file, err)
	}

	slug, ok := runtime.Slug(code)
	if !ok {
		return "", fmt.Errorf("cannot find a slug in %s", file)
	}

	return slug, nil
}
