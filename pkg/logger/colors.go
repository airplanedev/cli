package logger

import "github.com/fatih/color"

var (
	Gray   = color.New(color.FgHiBlack).SprintfFunc()
	Blue   = color.New(color.FgHiBlue).SprintfFunc()
	Red    = color.New(color.FgHiRed).SprintfFunc()
	Yellow = color.New(color.FgHiYellow).SprintfFunc()
	Green  = color.New(color.FgHiGreen).SprintfFunc()
	Bold   = color.New(color.Bold).SprintfFunc()
)
