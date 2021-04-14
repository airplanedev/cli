// Utilities for working with CLI inputs and API values
package execute

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/pkg/errors"
)

func promptFromParam(param api.Parameter) (survey.Prompt, error) {
	message := fmt.Sprintf("%s (%s):", param.Name, param.Slug)
	// TODO: support default values
	switch param.Type {
	case api.TypeString:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	case api.TypeBoolean:
		return &survey.Select{
			Message: message,
			Help:    param.Desc,
			Options: []string{"Yes", "No"},
		}, nil
	case api.TypeUpload:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	case api.TypeInteger:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	case api.TypeFloat:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	case api.TypeDate:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	case api.TypeDatetime:
		return &survey.Input{
			Message: message,
			Help:    param.Desc,
		}, nil
	default:
		return nil, errors.Errorf("unexpected param type %s", param.Type)
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
