package cli

import (
	"github.com/airplanedev/cli/pkg/api"
)

// Config represents command configuration.
//
// The config is passed down to all commands from
// the root command.
type Config struct {
	// Client represents the API client to use.
	//
	// It is initialized in the root command and passed
	// down to all commands.
	Client *api.Client

	// DebugMode indicates if the CLI should produce additional
	// debug output to guide end-users through issues.
	DebugMode bool

	// Version indicates if the CLI version should be printed.
	Version bool
}

// Must should be used for Cobra initialize commands that can return an error
// to enforce that they do not produce errors.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
