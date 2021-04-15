// Utilities for working with CLI inputs and API values
package execute

import (
	"fmt"
	"os"
	"reflect"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/params"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
)

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
			logger.Log("  %s%s %s", param.Name, req, logger.Gray("(--%s)", param.Slug))
			logger.Log("    %s %s", param.Type, param.Desc)
		}
		return errors.New("missing parameters")
	}

	logger.Log("You are about to run %s:", logger.Bold(task.Name))
	logger.Log(logger.Gray(client.TaskURL(task.ID)))
	logger.Log("")

	for _, param := range task.Parameters {
		if param.Type == api.TypeUpload {
			logger.Log(logger.Yellow("Skipping %s - uploads are not supported in CLI", param.Name))
			continue
		}

		prompt, err := promptForParam(param)
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

		value, err := params.ParseInput(param, inputValue)
		if err != nil {
			return err
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

// promptForParam returns a survey.Prompt matching the param type
func promptForParam(param api.Parameter) (survey.Prompt, error) {
	message := fmt.Sprintf("%s %s:", param.Name, logger.Gray("(--%s)", param.Slug))
	defaultValue, err := params.APIValueToInput(param, param.Default)
	if err != nil {
		return nil, err
	}
	switch param.Type {
	case api.TypeBoolean:
		return &survey.Select{
			Message: message,
			Help:    param.Desc,
			Options: []string{params.YesString, params.NoString},
			Default: defaultValue,
		}, nil
	default:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
			Default: defaultValue,
		}, nil
	}
}

// validateInput returns a survey.Validator to perform rudimentary checks on CLI input
func validateInput(param api.Parameter) func(interface{}) error {
	return func(ans interface{}) error {
		var v string
		switch a := ans.(type) {
		case string:
			v = a
		case survey.OptionAnswer:
			v = a.Value
		default:
			return errors.Errorf("unexpected answer of type %s", reflect.TypeOf(a).Name())
		}
		return params.ValidateInput(param, v)
	}
}
