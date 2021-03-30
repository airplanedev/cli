package logger

import (
	"fmt"
	"os"
)

var (
	// EnableDebug determines if debug logs are emitted.
	EnableDebug bool
)

// Log writes a log message to stderr, followed by a newline. Printf-style
// formatting is applied to msg using args.
func Log(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

// Debug writes a log message to stderr, followed by a newline, if the CLI
// is executing in debug mode. Printf-style formatting is applied to msg
// using args.
func Debug(msg string, args ...interface{}) {
	if EnableDebug {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
}
