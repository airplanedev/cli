package analytics

import (
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/getsentry/sentry-go"
	"gopkg.in/segmentio/analytics-go.v3"
)

var (
	segmentClient   analytics.Client
	segmentWriteKey string
	sentryDSN       string
)

func Init(debug bool) error {
	c, err := conf.ReadDefault()
	if err != nil {
		return err
	}
	if c.EnableTelemetry == "" {
		// User has not specified one way or the other, ask them to opt-in.
		if err := telemetryOptIn(c); err != nil {
			return err
		}
		// Now try again.
		return Init(debug)
	}
	if c.EnableTelemetry != "yes" {
		return nil
	}
	segmentClient = analytics.New(segmentWriteKey)
	return sentry.Init(sentry.ClientOptions{
		Dsn:   sentryDSN,
		Debug: debug,
	})
}

func telemetryOptIn(c conf.Config) error {
	var allow bool
	logger.Log("Welcome to the Airplane CLI!")
	logger.Log("")
	logger.Log("Is it OK for Airplane to collect usage analytics and error reports? This data will solely be used to improve the service.")
	logger.Log("")
	prompt := &survey.Confirm{
		Message: "Opt in",
		Default: true,
	}
	if err := survey.AskOne(
		prompt,
		&allow,
		survey.WithStdio(os.Stdin, os.Stderr, os.Stderr),
	); err != nil {
		return err
	}
	if allow {
		c.EnableTelemetry = "yes"
	} else {
		c.EnableTelemetry = "no"
	}
	if err := conf.WriteDefault(c); err != nil {
		return err
	}
	return nil
}

func Close() {
	if segmentClient != nil {
		if err := segmentClient.Close(); err != nil {
			logger.Debug("error closing segment client: %v", err)
		}
	}
	sentry.Flush(1 * time.Second)
}

type TrackOpts struct {
	UserID string
	TeamID string
	// Specify SkipSlack to avoid sending this event to Slack
	SkipSlack bool
}

// Track sends a track event to Segment.
// event should match "[event] by [user]" - e.g. "[Invite Sent] by [Alice]"
func Track(c *cli.Config, event string, properties map[string]interface{}) {
	ti := c.TokenInfo()
	props := analytics.NewProperties().Set("team_id", ti.TeamID)
	for k, v := range properties {
		props = props.Set(k, v)
	}
	if err := segmentClient.Enqueue(analytics.Track{
		UserId:     ti.UserID,
		Event:      event,
		Properties: props,
		Integrations: map[string]interface{}{
			"Slack": true,
		},
	}); err != nil {
		// Log but otherwise suppress the error
		sentry.CaptureException(err)
	}
}
