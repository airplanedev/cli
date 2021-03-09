package cli

import "github.com/airplanedev/cli/pkg/api"

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
}
