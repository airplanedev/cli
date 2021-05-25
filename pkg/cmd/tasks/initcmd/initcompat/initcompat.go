// Package initcompat provides the previous init command functionality.
package initcompat

import (
	"context"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/pkg/errors"
)

// Config configures the init command.
type Config struct {
	Client *api.Client
	From   string
	File   string
}

// Run runs the init command.
func Run(ctx context.Context, cfg Config) error {
	var kind initKind
	var err error
	// If --from is provided, we already know the user wants to create
	// from an existing task, so we don't need to prompt the user here.
	if cfg.From == "" {
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
		if err := initFromSample(ctx, cfg); err != nil {
			return err
		}
	case initKindScratch:
		if err := initFromScratch(ctx, cfg); err != nil {
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
