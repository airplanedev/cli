package initcmd

import (
	"context"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/ap"
	_ "github.com/airplanedev/cli/pkg/ap/django"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/pkg/errors"
)

// Project initializes a project.
func project(ctx context.Context, name string) error {
	root, err := filepath.Abs(".")
	if err != nil {
		return errors.Wrap(err, "filepath.abs(.)")
	}

	project, err := ap.ReadProject(root)

	if errors.Is(err, ap.ErrNoProject) {
		return createProject(name, root)
	}

	if err != nil {
		return err
	}

	logger.Log("Project file already exists with %d tasks.", len(project.Tasks))
	logger.Log("You can deploy all tasks using:")
	logger.Log("  airplane deploy airplane.json")
	return nil
}

// CreateProject creates a new project with framework name at root.
func createProject(name, root string) error {
	f, err := ap.LookupFramework(name, root)
	if err != nil {
		return err
	}

	cmds, err := f.ListCommands()
	if err != nil {
		return err
	}

	if len(cmds) == 0 {
		return errors.Errorf("could not find any commands")
	}

	cmds, err = selectCommands(cmds)
	if err != nil {
		return err
	}

	if len(cmds) == 0 {
		return errors.Errorf("you must select at least one command")
	}

	project := ap.NewProject(root)
	project.Framework = name
	project.Tasks = cmds
	if err := project.Save(); err != nil {
		return err
	}

	logger.Log("Saved project file with %d tasks", len(cmds))
	logger.Log("You can deploy all tasks using:")
	logger.Log("  airplane deploy airplane.json")
	return nil
}

// SelectCommands prompts the user to select commands.
func selectCommands(cmds []string) ([]string, error) {
	var selected []string
	var multi = survey.MultiSelect{
		Message: "Which commands would you like to add to Airplane?",
		Options: cmds,
	}

	if err := survey.AskOne(&multi, &selected); err != nil {
		return nil, err
	}

	return selected, nil
}
