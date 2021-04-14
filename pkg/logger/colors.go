package logger

import "github.com/fatih/color"

var (
	Gray   = color.New(color.FgHiBlack).SprintFunc()
	Blue   = color.New(color.FgHiBlue).SprintFunc()
	Red    = color.New(color.FgHiRed).SprintFunc()
	Yellow = color.New(color.FgHiYellow).SprintFunc()
	Green  = color.New(color.FgHiGreen).SprintFunc()
	Bold   = color.New(color.Bold).SprintFunc()
)
