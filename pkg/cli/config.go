package cli

import (
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
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

	// WithTelemetry indicates if the CLI should send usage analytics and errors, even if it's been
	// previously disabled.
	WithTelemetry bool

	// Version indicates if the CLI version should be printed.
	Version bool
}

// ParseTokenForAnalytics parses UNVERIFIED JWT information - this information can be spoofed.
// Should only be used for analytics, nothing sensitive.
func (c Config) ParseTokenForAnalytics() AnalyticsToken {
	var res AnalyticsToken
	token := c.Client.Token
	if token == "" {
		return res
	}
	t, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		logger.Debug("error parsing token: %v", err)
		return res
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return res
	}
	res.UserID = claims["userID"].(string)
	res.TeamID = claims["teamID"].(string)
	return res
}

type AnalyticsToken struct {
	UserID string
	TeamID string
}

// Must should be used for Cobra initialize commands that can return an error
// to enforce that they do not produce errors.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}
