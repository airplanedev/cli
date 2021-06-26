package logger

import (
	"fmt"
	"os"
	"strings"
)

var (
	// EnableDebug determines if debug logs are emitted.
	EnableDebug bool
)

// Log writes a log message to stderr, followed by a newline. Printf-style
// formatting is applied to msg using args.
func Log(msg string, args ...interface{}) {
	if len(args) == 0 {
		// Use Fprint if no args - avoids treating msg like a format string
		fmt.Fprint(os.Stderr, msg+"\n")
	} else {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
}

// Step prints a step that was performed.
func Step(msg string, args ...interface{}) {
	Log("- "+msg, args...)
}

// Suggest suggests a command with title and args.
func Suggest(title, command string, args ...interface{}) {
	Log("\n"+Gray(title)+"\n  "+command, args...)
}

// Error logs an error message.
func Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, Red("Error: ")+msg+"\n", args...)
}

// Warning logs a warning message.
func Warning(msg string, args ...interface{}) {
	fmt.Fprint(os.Stderr, Yellow("[warning] "+msg+"\n", args...))
}

// Debug writes a log message to stderr, followed by a newline, if the CLI
// is executing in debug mode. Printf-style formatting is applied to msg
// using args.
func Debug(msg string, args ...interface{}) {
	if !EnableDebug {
		return
	}

	msgf := msg
	if len(args) > 0 {
		msgf = fmt.Sprintf(msg, args...)
	}

	debugPrefix := "[" + Blue("debug") + "] "
	msgf = debugPrefix + strings.Join(strings.Split(msgf, "\n"), "\n"+debugPrefix)

	fmt.Fprint(os.Stderr, msgf+"\n")
}
