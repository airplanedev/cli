package configs

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/pkg/errors"
)

// ReadValue reads a config value from prompt if allowed, else stdin
func ReadValue(secret bool) (string, error) {
	if utils.CanPrompt() {
		return ReadValueFromPrompt("Config value:", secret)
	}
	// Read from stdin
	logger.Log("Reading secret from stdin...")
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", errors.Wrap(err, "reading from stdin")
	}
	return strings.TrimSpace(string(data)), nil
}

// ReadValueFromPrompt prompts user for config value
func ReadValueFromPrompt(message string, secret bool) (string, error) {
	var value string
	var prompt survey.Prompt
	if secret {
		prompt = &survey.Password{Message: message}
	} else {
		prompt = &survey.Input{Message: message}
	}
	if err := survey.AskOne(
		prompt,
		&value,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return "", errors.Wrap(err, "prompting value")
	}
	return strings.TrimSpace(value), nil
}

// SetConfig writes config value to API and prints progress to user
func SetConfig(ctx context.Context, client *api.Client, nt NameTag, value string, secret bool) error {
	// Avoid printing back secrets
	var valueStr string
	if secret {
		valueStr = "<secret value>"
	} else {
		valueStr = value
	}
	logger.Log("  Setting %s to %s...", logger.Blue(JoinName(nt)), logger.Green(valueStr))
	req := api.SetConfigRequest{
		Name:     nt.Name,
		Tag:      nt.Tag,
		Value:    value,
		IsSecret: secret,
	}
	if err := client.SetConfig(ctx, req); err != nil {
		return errors.Wrap(err, "set config")
	}
	logger.Log("  Done!")
	return nil
}
