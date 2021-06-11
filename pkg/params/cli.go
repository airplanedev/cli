package params

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
)

// CLI parses a list of flags as Airplane parameters and returns the values.
//
// A flag.ErrHelp error will be returned if a -h or --help was provided, in which case
// this function will print out help text on how to pass this task's parameters as flags.
func CLI(args []string, client *api.Client, task api.Task) (api.Values, error) {
	values := api.Values{}

	if len(args) > 0 {
		// If args have been passed in, parse them as flags
		set := flagset(task, values)
		if err := set.Parse(args); err != nil {
			return nil, err
		}
	} else {
		// Otherwise, try to prompt for parameters
		if err := promptForParamValues(client, task, values); err != nil {
			return nil, err
		}
	}

	return values, nil
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
			args[p.Slug], err = ParseInput(p, v)
			if err != nil {
				return errors.Wrap(err, "converting input to API value")
			}
			return
		})
	}

	return set
}

// promptForParamValues attempts to prompt user for param values, setting them on `params`
// If there are no parameters, does nothing.
// If TTY, prompts for parameters and then asks user to confirm.
// If no TTY, errors.
func promptForParamValues(client *api.Client, task api.Task, paramValues map[string]interface{}) error {
	if len(task.Parameters) == 0 {
		return nil
	}

	if !utils.CanPrompt() {
		// Error since we have no params and no way to prompt for it
		// TODO: if all parameters optional (or have defaults), do not error.
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
		if param.Constraints.Regex != "" {
			opts = append(opts, survey.WithValidator(regexValidator(param.Constraints.Regex)))
		}
		var inputValue string
		if err := survey.AskOne(prompt, &inputValue, opts...); err != nil {
			return errors.Wrap(err, "asking prompt for param")
		}

		value, err := ParseInput(param, inputValue)
		if err != nil {
			return err
		}
		if value != nil {
			paramValues[param.Slug] = value
		}
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
	defaultValue, err := APIValueToInput(param, param.Default)
	if err != nil {
		return nil, err
	}
	switch param.Type {
	case api.TypeBoolean:
		var dv interface{}
		if defaultValue == "" {
			dv = nil
		} else {
			dv = defaultValue
		}
		return &survey.Select{
			Message: message,
			Help:    param.Desc,
			Options: []string{YesString, NoString},
			Default: dv,
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
		return ValidateInput(param, v)
	}
}

// regexValidator returns a Survey validator from the pattern
func regexValidator(pattern string) func(interface{}) error {
	return func(val interface{}) error {
		str, ok := val.(string)
		if !ok {
			return errors.New("expected string")
		}
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			return errors.Errorf("errored matching against regex: %s", err)
		}
		if !matched {
			return errors.Errorf("must match regex pattern: %s", pattern)
		}
		return nil
	}
}
