package analytics

import (
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
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

func Init() error {
	c, err := conf.ReadDefault()
	if err != nil {
		return err
	}
	if c.EnableTelemetry == "" {
		var allow bool
		logger.Log("Is it OK to collect usage analytics and error reports? This data will solely be used to improve Airplane.")
		logger.Log("")
		prompt := &survey.Confirm{
			Message: "Allow analytics and error reporting?",
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
		return Init()
	}
	if c.EnableTelemetry != "yes" {
		return nil
	}
	segmentClient = analytics.New(segmentWriteKey)
	return sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
	})
}

func Close() {
	if segmentClient != nil {
		if err := segmentClient.Close(); err != nil {
			logger.Debug("error closing segment client: %v", err)
		}
	}
	sentry.Flush(1 * time.Second)
}

func enqueue(msg analytics.Message) {
	if err := segmentClient.Enqueue(msg); err != nil {
		// Log but otherwise suppress the error
		sentry.CaptureException(err)
	}
}

type TrackOpts struct {
	UserID string
	TeamID string
	// Specify SkipSlack to avoid sending this event to Slack
	SkipSlack bool
}

// Track sends a track event to Segment.
// event should match "[event] by [user]" - e.g. "[Invite Sent] by [Alice]"
func Track(userID string, event string, properties map[string]interface{}) {
	props := analytics.NewProperties()
	for k, v := range properties {
		props = props.Set(k, v)
	}
	enqueue(analytics.Track{
		UserId:     userID,
		Event:      event,
		Properties: props,
		Integrations: map[string]interface{}{
			"Slack": true,
		},
	})
}
