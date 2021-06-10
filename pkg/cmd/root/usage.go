package root

import (
	"fmt"
	"strings"

	"github.com/airplanedev/cli/pkg/logger"
	"github.com/kr/text"
	"github.com/spf13/cobra"
)

// Usage prints the usage for a command.
func usage(cmd *cobra.Command) error {
	return nil
}

// Help prints the help for a command.
func help(cmd *cobra.Command, args []string) {
	desc := cmd.Short
	if cmd.Long != "" {
		desc = cmd.Long
	}
	logger.Log("%s", desc)
	logger.Log("")
	logger.Log("%s", logger.Bold("Usage:"))
	logger.Log("  %s", cmd.UseLine())

	if cmd.HasSubCommands() {
		logger.Log("\n%s", logger.Bold("Commands:"))
		for _, cmd := range cmd.Commands() {
			if !cmd.Hidden {
				name := rpad(cmd.Name(), cmd.NamePadding())
				logger.Log("  %s", name+cmd.Short)
			}
		}
	}

	if flags := cmd.LocalFlags().FlagUsages(); flags != "" {
		s := dedent(flags)
		logger.Log("\n%s", logger.Bold("Flags:"))
		logger.Log("%s", text.Indent(s, "  "))
	}

	if cmd.HasExample() {
		s := trim(cmd.Example)
		logger.Log("\n%s", logger.Bold("Examples:"))
		logger.Log("%s", text.Indent(s, "  "))
	}

	logger.Log("")
}

// Trim trims all spaces.
func trim(s string) string {
	return strings.TrimSpace(s)
}

// Dedent trims spaces from each line.
func dedent(s string) string {
	var lines = strings.Split(s, "\n")
	var ret []string
	var min = -1

	for _, line := range lines {
		if len(line) > 0 {
			indent := len(line) - len(strings.TrimLeft(line, " "))
			if min == -1 || indent < min {
				min = indent
			}
		}
	}

	for _, l := range lines {
		ret = append(ret, strings.TrimPrefix(l,
			strings.Repeat(" ", min),
		))
	}

	return strings.Join(ret, "\n")
}

// Rpad rpads the given string.
func rpad(s string, n int) string {
	t := fmt.Sprintf("%%-%ds", n)
	return fmt.Sprintf(t, s)
}
