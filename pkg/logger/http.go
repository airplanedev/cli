package logger

import (
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

// HTTPLogger is a wrapper around the default debug logger that
// can be used for HTTP requests via `hashicorp/go-retryablehttp`.
type HTTPLogger struct{}

var _ retryablehttp.LeveledLogger = HTTPLogger{}

func (_ HTTPLogger) Error(msg string, keyAndValues ...interface{}) {
	format, args := toFormatAndArgs(msg, keyAndValues)
	Debug(format, args...)
}

func (_ HTTPLogger) Info(msg string, keyAndValues ...interface{}) {
	format, args := toFormatAndArgs(msg, keyAndValues)
	Debug(format, args...)
}

func (_ HTTPLogger) Debug(msg string, keyAndValues ...interface{}) {
	format, args := toFormatAndArgs(msg, keyAndValues)
	Debug(format, args...)
}

func (_ HTTPLogger) Warn(msg string, keyAndValues ...interface{}) {
	format, args := toFormatAndArgs(msg, keyAndValues)
	Debug(format, args...)
}

func toFormatAndArgs(msg string, keyAndValues []interface{}) (string, []interface{}) {
	var keys []string
	var args []interface{}
	for i, v := range keyAndValues {
		// keyAndValues is a list of format: [key, value, key, value, ...]
		//
		// Every time we see a value, add (key, value) to args.
		if i%2 == 0 {
			continue
		}

		if _, ok := keyAndValues[i-1].(string); ok {
			keys = append(keys, "%v")
			args = append(args, v)
		}
	}

	format := msg
	if len(keys) > 0 {
		format += " " + strings.Join(keys, " ")
	}

	return format, args
}
