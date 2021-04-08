// initcmd defines the implementation of the `airplane tasks init` command.
//
// Even though the command is called "init", we can't name the package "init"
// since that conflicts with the Go init function.
package initcmd

import (
	"context"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	root *cli.Config
	file string
	from string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{root: c}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a task definition",
		Example: heredoc.Doc(`
			$ airplane tasks init
			$ airplane tasks init -f ./airplane.yml
			$ airplane tasks init --from hello_world
		`),
		PersistentPreRunE: utils.WithParentPersistentPreRunE(func(cmd *cobra.Command, args []string) error {
			return login.EnsureLoggedIn(cmd.Root().Context(), c)
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Root().Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a file to store task definition")
	cmd.Flags().StringVar(&cfg.from, "from", "", "Slug of an existing task to generate from")

	return cmd
}

func run(ctx context.Context, cfg config) error {
	var kind initKind
	var err error
	// If --from is provided, we already know the user wants to create
	// from an existing task, so we don't need to prompt the user here.
	if cfg.from == "" {
		logger.Log("Airplane is a development platform for engineers building internal tools.\n")
		logger.Log("This command will configure a task definition which Airplane uses to deploy your task.\n")

		if kind, err = pickInitKind(); err != nil {
			return err
		}
	} else {
		kind = initKindTask
	}

	switch kind {
	case initKindSample:
		if err := initFromSample(cfg); err != nil {
			return err
		}
	case initKindScratch:
		if err := initFromScratch(cfg); err != nil {
			return err
		}
	case initKindTask:
		if err := initFromTask(ctx, cfg); err != nil {
			return err
		}
	default:
		return errors.Errorf("Unexpected unknown initKind choice: %s", kind)
	}

	return nil
}

type initKind string

const (
	initKindSample  initKind = "Create from an Airplane-provided sample"
	initKindScratch initKind = "Create from scratch"
	initKindTask    initKind = "Import from an existing Airplane task"
)

func pickInitKind() (initKind, error) {
	var kind string
	if err := survey.AskOne(
		&survey.Select{
			Message: "How do you want to get started?",
			// TODO: disable the search filter on this Select. Will require an upstream
			// change to the survey repo.
			Options: []string{
				string(initKindSample),
				string(initKindScratch),
				string(initKindTask),
			},
			Default: string(initKindSample),
		},
		&kind,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return initKind(""), errors.Wrap(err, "selecting kind of init")
	}

	return initKind(kind), nil
}
