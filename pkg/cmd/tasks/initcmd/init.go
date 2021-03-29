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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	cli  *cli.Config
	file string
	from string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{cli: c}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a task definition",
		Example: heredoc.Doc(`
			$ airplane tasks init
			$ airplane tasks init -f ./airplane.yml
			$ airplane tasks init --from hello_world
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cmd, cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "", "Path to a file to store task definition")
	cmd.Flags().StringVar(&cfg.from, "from", "", "Slug of an existing task to generate from")

	return cmd
}

func run(ctx context.Context, cmd *cobra.Command, cfg config) error {
	var kind initKind
	var err error
	// If --from is provided, we already know the user wants to create
	// from an existing task, so we don't need to prompt the user here.
	if cfg.from == "" {
		cmd.Printf("Airplane is the code-first solution for engineers building internal tools.\n\n")
		cmd.Printf("This command will configure a task definition which specifies how to deploy your task to Airplane.\n\n")

		if kind, err = pickInitKind(); err != nil {
			return err
		}
	}

	switch kind {
	case initKindSample:
		if err := initFromSample(cmd, cfg); err != nil {
			return err
		}
	case initKindScratch:
		if err := initFromScratch(cmd, cfg); err != nil {
			return err
		}
	case initKindExisting:
		if err := initFromExisting(ctx, cmd, cfg); err != nil {
			return err
		}
	default:
		return errors.Errorf("Unexpected unknown initKind choice: %s", kind)
	}

	return nil
}

type initKind string

var (
	initKindSample   initKind = "Create from an Airplane-provided sample"
	initKindScratch  initKind = "Create from scratch"
	initKindExisting initKind = "Create from an existing Airplane task"
)

func pickInitKind() (initKind, error) {
	var kind string
	if err := survey.AskOne(
		&survey.Select{
			Message: "How do you want to get started?",
			// TODO: upstream the ability to disable Survey's search filter
			Options: []string{
				string(initKindSample),
				string(initKindScratch),
				string(initKindExisting),
			},
			Default: string(initKindSample),
			// Help:    "todo",
		},
		&kind,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return initKind(""), errors.Wrap(err, "selecting kind of init")
	}

	return initKind(kind), nil
}
