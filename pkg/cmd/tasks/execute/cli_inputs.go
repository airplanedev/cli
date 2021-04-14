// Utilities for working with CLI inputs and API values
package execute

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
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
			logger.Log("  %s%s (%s)", param.Slug, req, param.Name)
			logger.Log(logger.Gray("    %s %s", param.Type, param.Desc))
		}
		return errors.New("missing parameters")
	}

	logger.Log("You are about to run %s:", logger.Bold(task.Name))
	logger.Log(logger.Gray(client.TaskURL(task.ID)))
	logger.Log("")

	for _, param := range task.Parameters {
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

// promptForParam returns a survey.Prompt matching the param type
func promptForParam(param api.Parameter) (survey.Prompt, error) {
	message := fmt.Sprintf("%s (%s):", param.Name, param.Slug)
	// TODO: support default values
	switch param.Type {
	case api.TypeBoolean:
		return &survey.Select{
			Message: message,
			Help:    param.Desc,
			Options: []string{"Yes", "No"},
		}, nil
	default:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	}
}

// Converts an inputted text value to the API value
// For booleans, this means something like "yes" becomes true
// For datetimes, this means the string remains the same (since the API still expects a string)
func inputToAPIValue(param api.Parameter, v string) (interface{}, error) {
	if v == "" {
		return param.Default, nil
	}
	switch param.Type {
	case api.TypeString, api.TypeDate, api.TypeDatetime:
		return v, nil

	case api.TypeBoolean:
		return parseBool(v)

	case api.TypeInteger:
		return strconv.Atoi(v)

	case api.TypeFloat:
		return strconv.ParseFloat(v, 64)

	case api.TypeUpload:
		if v != "" {
			return nil, errors.New("uploads are not supported from the CLI")
		}
		return nil, nil

	default:
		return v, nil
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

		// Treat empty value as valid - optional/required is checked separately.
		if v == "" {
			return nil
		}

		switch param.Type {
		case api.TypeString:
			return nil

		case api.TypeBoolean:
			if _, err := parseBool(v); err != nil {
				return errors.New("expected yes, no, true, false, 1 or 0")
			}

		case api.TypeInteger:
			if _, err := strconv.Atoi(v); err != nil {
				return errors.New("invalid integer")
			}

		case api.TypeFloat:
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return errors.New("invalid number")
			}

		case api.TypeUpload:
			if v != "" {
				// TODO(amir): we need to support them with some special
				// character perhaps `@` like curl?
				return errors.New("uploads are not supported from the CLI")
			}

		case api.TypeDate:
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return errors.New("expected to be formatted as '2016-01-02'")
			}
		case api.TypeDatetime:
			if _, err := time.Parse("2006-01-02T15:04:05Z", v); err != nil {
				return errors.New("expected to be formatted as '2016-01-02T15:04:05Z'")
			}
			return nil
		}
		return nil
	}
}

// Light wrapper around strconv.ParseBool with support for yes and no
func parseBool(v string) (bool, error) {
	switch v {
	case "Yes", "yes":
		return true, nil
	case "No", "no":
		return false, nil
	default:
		return strconv.ParseBool(v)
	}
}
