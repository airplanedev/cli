package logger

import (
	"fmt"
	"os"
)

var (
	EnableDebug bool
)

func Log(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
}

func Debug(msg string, args ...interface{}) {
	if EnableDebug {
		fmt.Fprintf(os.Stderr, msg, args...)
	}
}
