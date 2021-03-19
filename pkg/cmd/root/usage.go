package root

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/kr/text"
	"github.com/spf13/cobra"
)

var (
	bold = color.New(color.Bold).SprintFunc()
)

// Usage prints the usage for a command.
func usage(cmd *cobra.Command) error {
	return nil
}

// Help prints the help for a command.
func help(cmd *cobra.Command, args []string) {
	cmd.Println()
	cmd.Printf("%s\n", bold("Usage:"))
	cmd.Printf("  %s\n", cmd.UseLine())

	if cmd.HasSubCommands() {
		cmd.Printf("\n%s\n", bold("Commands:"))
		for _, cmd := range cmd.Commands() {
			if !cmd.Hidden {
				name := rpad(cmd.Name(), cmd.NamePadding())
				cmd.Printf("  %s\n", name+cmd.Short)
			}
		}
	}

	if flags := cmd.LocalFlags().FlagUsages(); flags != "" {
		s := trim(dedent(flags))
		cmd.Printf("\n%s\n", bold("Flags:"))
		cmd.Printf("%s\n", text.Indent(s, "  "))
	}

	if cmd.HasExample() {
		s := trim(cmd.Example)
		cmd.Printf("\n%s\n", bold("Examples:"))
		cmd.Printf("%s\n", text.Indent(s, "  "))
	}

	cmd.Println()
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
