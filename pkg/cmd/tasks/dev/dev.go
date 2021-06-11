package dev

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/outputs"
	"github.com/airplanedev/cli/pkg/params"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type config struct {
	root *cli.Config
	file string
	args []string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}

	cmd := &cobra.Command{
		Use:   "dev ./path/to/file",
		Short: "Locally run a task",
		Long:  "Locally runs a task, optionally with specific parameters.",
		Example: heredoc.Doc(`
			airplane dev ./task.js [-- <parameters...>]
			airplane dev ./task.ts [-- <parameters...>]
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			// TODO: update the `dev` command to work w/out internet access
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

func run(ctx context.Context, cfg config) error {
	if !fsx.Exists(cfg.file) {
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

	logger.Log("Locally running %s task %s", logger.Bold(task.Name), logger.Gray("("+cfg.root.Client.TaskURL(task.Slug)+")"))
	logger.Log("")

	cmds, err := r.PrepareRun(ctx, runtime.PrepareRunOptions{
		Path:        cfg.file,
		ParamValues: paramValues,
		KindOptions: task.KindOptions,
	})
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, cmds[0], cmds[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "stderr")
	}
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "starting")
	}

	// mu guards o
	var mu sync.Mutex
	o := api.Outputs{}

	logParser := func(r io.Reader) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if outputs.IsOutput(line) {
				name := outputs.ParseOutputName(line)
				value := outputs.ParseOutputValue(line)

				mu.Lock()
				o[name] = append(o[name], value)
				mu.Unlock()
			}
			logger.Log("[%s] %s", logger.Gray("log"), line)
		}
		return errors.Wrap(scanner.Err(), "scanning logs")
	}

	eg := errgroup.Group{}
	eg.Go(func() error {
		return logParser(stdout)
	})
	eg.Go(func() error {
		return logParser(stderr)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "waiting")
	}

	print.Outputs(o)

	return nil
}

// slugFromScript attempts to extract a slug from a script.
func slugFromScript(file string) (string, error) {
	code, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s - %w", file, err)
	}

	slug, ok := runtime.Slug(code)
	if !ok {
		return "", runtime.ErrNotLinked{Path: file}
	}

	return slug, nil
}
