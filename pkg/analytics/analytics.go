package analytics

import (
	"time"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/utils"
	"github.com/airplanedev/cli/pkg/version"
	"github.com/getsentry/sentry-go"
	"gopkg.in/segmentio/analytics-go.v3"
)

// Set by Go Releaser.
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
	if c.EnableTelemetry == nil {
		// User has not specified one way or the other, ask them to opt-in.
		if err := telemetryOptIn(c); err != nil {
			return err
		}
		// Now try again.
		return Init(debug)
	}
	if !*c.EnableTelemetry {
		return nil
	}
	segmentClient = analytics.New(segmentWriteKey)
	return sentry.Init(sentry.ClientOptions{
		Dsn:     sentryDSN,
		Debug:   debug,
		Release: version.Get(),
	})
}

func telemetryOptIn(c conf.Config) error {
	var allow bool
	logger.Log("Is it OK for Airplane to collect usage analytics and error reports? This data will solely be used to improve the service.")
	logger.Log("")
	allow, err := utils.Confirm("Opt in")
	if err != nil {
		return err
	}
	c.EnableTelemetry = &allow
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
	if segmentClient == nil {
		return
	}
	tok := c.ParseTokenForAnalytics()
	props := analytics.NewProperties().Set("team_id", tok.TeamID)
	for k, v := range properties {
		props = props.Set(k, v)
	}
	enqueue(analytics.Track{
		UserId:     tok.UserID,
		Event:      event,
		Properties: props,
		Integrations: map[string]interface{}{
			"Slack": true,
		},
	})
}

func enqueue(msg analytics.Message) {
	if segmentClient == nil {
		return
	}
	if err := segmentClient.Enqueue(msg); err != nil {
		// Log (but otherwise suppress) the error
		sentry.CaptureException(err)
	}
}
