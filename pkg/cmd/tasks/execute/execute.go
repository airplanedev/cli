package execute

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/fs"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/params"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/airplanedev/cli/pkg/runtime"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Config is the execute config.
type config struct {
	root *cli.Config
	task string // Could be a file, yaml definition or a slug.
	args []string
}

// New returns a new execute cobra command.
func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}

	cmd := &cobra.Command{
		Use:     "execute <slug>",
		Short:   "Execute a task",
		Aliases: []string{"exec"},
		Long:    "Execute a task from the CLI, optionally with specific parameters.",
		Example: heredoc.Doc(`
			airplane execute ./task.js [-- <parameters...>]
			airplane execute hello_world [-- <parameters...>]
			airplane execute ./airplane.yml [-- <parameters...>]
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.task != "" {
				// A file was provided with the -f flag. This is deprecated.
				logger.Warning(`The --file/-f flag is deprecated and will be removed in a future release. File paths should be passed as a positional argument instead: airplane execute %s`, cfg.task)
				cfg.args = args
			} else if len(args) > 0 {
				cfg.task = args[0]
				cfg.args = args[1:]
			} else {
				return errors.New("expected 1 argument: airplane execute [./path/to/file | task slug]")
			}

			return run(cmd.Root().Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.task, "file", "f", "", "File to deploy (.yaml, .yml, .js, .ts)")
	cli.Must(cmd.Flags().MarkHidden("file")) // --file is deprecated

	return cmd
}

// Run runs the execute command.
func run(ctx context.Context, cfg config) error {
	var client = cfg.root.Client

	// cfg.task is either a slug or a local path. Try it as a slug first.
	task, err := client.GetTask(ctx, cfg.task)
	if _, ok := err.(*api.TaskMissingError); ok {
		// If there's no task matching that slug, try it as a file path instead.
		if !fs.Exists(cfg.task) {
			return errors.Errorf("Unable to execute %s. No matching file or task slug.", cfg.task)
		}

		slug, err := slugFrom(cfg.task)
		if err != nil {
			return err
		}

		task, err = client.GetTask(ctx, slug)
		if err != nil {
			return errors.Wrap(err, "get task")
		}
	} else if err != nil {
		return errors.Wrap(err, "get task")
	}

	if task.Image == nil {
		return &notDeployedError{
			task: cfg.task,
		}
	}

	req := api.RunTaskRequest{
		TaskID:      task.ID,
		ParamValues: make(api.Values),
	}

	if len(cfg.args) > 0 {
		// If args have been passed in, parse them as flags
		set := flagset(task, req.ParamValues)
		if err := set.Parse(cfg.args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
	} else {
		// Otherwise, try to prompt for parameters
		if err := promptForParamValues(cfg.root.Client, task, req.ParamValues); err != nil {
			return err
		}
	}

	logger.Log(logger.Gray("Running: %s", task.Name))

	w, err := client.Watcher(ctx, req)
	if err != nil {
		return err
	}

	logger.Log(logger.Gray("Queued: %s", client.RunURL(w.RunID())))

	var state api.RunState
	agentPrefix := "[agent]"
	outputPrefix := "airplane_output"

	for {
		if state = w.Next(); state.Err() != nil {
			break
		}

		for _, l := range state.Logs {
			var loggedText string
			if strings.HasPrefix(l.Text, agentPrefix) {
				// De-emphasize agent logs and remove prefix
				loggedText = logger.Gray(strings.TrimLeft(strings.TrimPrefix(l.Text, agentPrefix), " "))
			} else if strings.HasPrefix(l.Text, outputPrefix) {
				// De-emphasize outputs appearing in logs
				loggedText = logger.Gray(l.Text)
			} else {
				// Try to leave user logs alone, so they can apply their own colors
				loggedText = l.Text
			}
			logger.Log(loggedText)
		}

		if state.Stopped() {
			break
		}
	}

	if err := state.Err(); err != nil {
		return err
	}

	print.Outputs(state.Outputs)

	status := string(state.Status)
	switch state.Status {
	case api.RunSucceeded:
		status = logger.Green(status)
	case api.RunFailed, api.RunCancelled:
		status = logger.Red(status)
	}
	logger.Log(logger.Bold(status))

	if state.Failed() {
		return errors.New("Run has failed")
	}

	return nil
}

// Flagset returns a new flagset from the given task parameters.
func flagset(task api.Task, args api.Values) *flag.FlagSet {
	var set = flag.NewFlagSet(task.Name, flag.ContinueOnError)

	set.Usage = func() {
		logger.Log("\n%s Usage:", task.Name)
		set.VisitAll(func(f *flag.Flag) {
			logger.Log("  --%s %s (default: %q)", f.Name, f.Usage, f.DefValue)
		})
		logger.Log("")
	}

	for i := range task.Parameters {
		// Scope p here (& not above) so we can use it in the closure.
		// See also: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		p := task.Parameters[i]
		set.Func(p.Slug, p.Desc, func(v string) (err error) {
			args[p.Slug], err = params.ParseInput(p, v)
			if err != nil {
				return errors.Wrap(err, "converting input to API value")
			}
			return
		})
	}

	return set
}

// SlugFrom returns the slug from the given file.
func slugFrom(file string) (string, error) {
	switch ext := filepath.Ext(file); ext {
	case ".yml", ".yaml":
		return slugFromYaml(file)
	case ".js", ".ts":
		return slugFromScript(file)
	case "":
		return "", fmt.Errorf("the file %s must have an extension", file)
	default:
		return "", fmt.Errorf("the file %s has unrecognized extension", file)
	}
}

// slugFromYaml attempts to extract a slug from a yaml definition.
func slugFromYaml(file string) (string, error) {
	dir, err := taskdir.Open(file)
	if err != nil {
		return "", err
	}
	defer dir.Close()

	def, err := dir.ReadDefinition()
	if err != nil {
		return "", err
	}

	if def.Slug == "" {
		return "", errors.Errorf("no task slug found in task definition at %s", file)
	}

	return def.Slug, nil
}

// slugFromScript attempts to extract a slug from a script.
func slugFromScript(file string) (string, error) {
	r, ok := runtime.Lookup(file)
	if !ok {
		return "", fmt.Errorf("%s tasks are not supported", file)
	}

	code, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s - %w", file, err)
	}

	slug, ok := r.Slug(code)
	if !ok {
		return "", fmt.Errorf("cannot find a slug in %s", file)
	}

	return slug, nil
}

type notDeployedError struct {
	task string
}

// Error implementation.
func (err notDeployedError) Error() string {
	return fmt.Sprintf("task %s was not deployed", err.task)
}

// ExplainError implementation.
func (err notDeployedError) ExplainError() string {
	return fmt.Sprintf("to deploy the task:\n  airplane deploy %s", err.task)
}
