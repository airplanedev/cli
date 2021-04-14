package logger

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	Gray   = ColorSprintfFunc(color.FgHiBlack)
	Blue   = ColorSprintfFunc(color.FgHiBlue)
	Red    = ColorSprintfFunc(color.FgHiRed)
	Yellow = ColorSprintfFunc(color.FgHiYellow)
	Green  = ColorSprintfFunc(color.FgHiGreen)
	Bold   = ColorSprintfFunc(color.Bold)
)

func ColorSprintfFunc(c color.Attribute) func(msg string, args ...interface{}) string {
	colorSprint := color.New(c).SprintFunc()
	return func(msg string, args ...interface{}) string {
		return colorSprint(fmt.Sprintf(msg, args...))
	}
}
