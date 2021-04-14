package execute

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	bold = color.New(color.Bold).SprintfFunc()
	gray = color.New(color.FgHiBlack).SprintfFunc()
)

// Config is the execute config.
type config struct {
	root *cli.Config
	slug string
	args []string
	file string
}

// New returns a new execute cobra command.
func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}

	cmd := &cobra.Command{
		Use:   "execute <slug>",
		Short: "Execute a task",
		Long:  "Execute a task by its slug with the provided parameters.",
		Example: heredoc.Doc(`
			airplane execute -f ./airplane.yml [-- <parameters...>]
			airplane execute hello_world [-- <parameters...>]
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := cmd.Flags().ArgsLenAtDash()
			if n > 1 {
				return errors.Errorf("at most one arg expected, got: %d", n)
			}

			// If a '--' was used, then we have 0 or more args to pass to the task.
			if n != -1 {
				cfg.args = args[n:]
			}

			// If an arg was passed, before the --, then it is a task slug to execute.
			if len(args) > 0 && n != 0 {
				cfg.slug = args[0]
			}

			return run(cmd.Root().Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a task definition file.")

	return cmd
}

// Run runs the execute command.
func run(ctx context.Context, cfg config) error {
	var client = cfg.root.Client

	slug := cfg.slug
	if slug == "" {
		if cfg.file == "" {
			return errors.New("expected either a task slug or --file")
		}

		dir, err := taskdir.Open(cfg.file)
		if err != nil {
			return err
		}
		defer dir.Close()

		def, err := dir.ReadDefinition()
		if err != nil {
			return err
		}

		if def.Slug == "" {
			return errors.Errorf("no task slug found in task definition at %s", cfg.file)
		}

		slug = def.Slug
	}

	task, err := client.GetTask(ctx, slug)
	if err != nil {
		return errors.Wrap(err, "get task")
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

	logger.Log(gray("Running: %s\n", task.Name))

	w, err := client.Watcher(ctx, req)
	if err != nil {
		return err
	}

	logger.Log(gray("Queued: %s", client.RunURL(w.RunID())))

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
				loggedText = gray(strings.TrimLeft(strings.TrimPrefix(l.Text, agentPrefix), " "))
			} else if strings.HasPrefix(l.Text, outputPrefix) {
				// De-emphasize outputs appearing in logs
				loggedText = gray(l.Text)
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
		status = color.GreenString(status)
	case api.RunFailed, api.RunCancelled:
		status = color.RedString(status)
	}
	logger.Log(bold(status))

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

	for _, p := range task.Parameters {
		set.Func(p.Slug, p.Desc, func(v string) (err error) {
			// TODO: refactor out this function to re-use for prompting as well
			args[p.Slug], err = inputToAPIValue(p, v)
			if err != nil {
				return errors.Wrap(err, "converting input to API value")
			}
			return
		})
	}

	return set
}

// promptForParamValues attempts to prompt user for param values, setting them on `params`
// If no TTY, errors unless there are no parameters
// If TTY, prompts for parameters (if any) and asks user to confirm
func promptForParamValues(client *api.Client, task api.Task, paramValues map[string]interface{}) error {
	if !utils.CanPrompt() {
		// Don't error if there are no params
		if len(task.Parameters) == 0 {
			return nil
		}
		// Otherwise, error since we have no params and no way to prompt for it
		logger.Log("Parameters were not specified! Task has %d parameter(s):\n", len(task.Parameters))
		for _, param := range task.Parameters {
			var req string
			if !param.Constraints.Optional {
				req = "*"
			}
			logger.Log("  %s%s %s (%s)", param.Name, req, param.Type, param.Slug)
			if param.Desc != "" {
				logger.Log("    %s", param.Desc)
			}
		}
		return errors.New("missing parameters")
	}

	logger.Log("You are about to run %s:", bold(task.Name))
	logger.Log(gray(client.TaskURL(task.ID)))
	logger.Log("")

	for _, param := range task.Parameters {
		prompt, err := promptFromParam(param)
		if err != nil {
			return err
		}
		opts := []survey.AskOpt{
			survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
			survey.WithValidator(validateInput(param)),
		}
		if !param.Constraints.Optional {
			opts = append(opts, survey.WithValidator(survey.Required))
		}
		var inputValue string
		if err := survey.AskOne(prompt, &inputValue, opts...); err != nil {
			return errors.Wrap(err, "asking prompt for param")
		}

		value, err := inputToAPIValue(param, inputValue)
		if err != nil {
			return errors.Wrap(err, "converting input to API value")
		}
		paramValues[param.Slug] = value
	}
	confirmed := false
	if err := survey.AskOne(&survey.Confirm{
		Message: "Execute?",
		Default: true,
	}, &confirmed); err != nil {
		return errors.Wrap(err, "confirming")
	}
	if !confirmed {
		return errors.New("user cancelled")
	}
	return nil
}
