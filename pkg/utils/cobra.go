package utils

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// WithParentPersistentPreRunE runs the parent command's PersistentPreRunE before the current
// command's PersistentPreRunE. This prevents the default Cobra behavior of only running the
// final PersistentPreRunE.
func WithParentPersistentPreRunE(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for parent := cmd.Parent(); parent != nil; {
			// Find the first parent with a PersistentPreRunE, if any.
			if parent.PersistentPreRunE == nil {
				continue
			}

			if err := parent.PersistentPreRunE(parent, args); err != nil {
				return err
			}
			break
		}

		return f(cmd, args)
	}
}

// TimeValue is a pflag.Value that can be used to parse a time.Time
// as a Cobra flag.
//
// For example:
//   var tv timeValue
//   cmd.Flags().Var(&tv, "since", "Filters by created_at")
//
// Which could be set as: `--since="2020-01-02T01:02:03"`
//
// TimeValue's are alias types of time.Time. You can convert safely via `time.Time(tv)`.
type TimeValue time.Time

var _ pflag.Value = &TimeValue{}

func (tv *TimeValue) Set(s string) error {
	// Cobra doesn't appear to support quoted strings with spaces:
	// https://github.com/spf13/cobra/issues/1114
	// If fixed, we could start supporting time formats with spaces like "2006-01-02 15:04:05".
	for _, format := range []string{
		// Overall, we are roughly looking for RFC3339 timestamps with some leeway
		// to make timestamps easier to specify.
		//
		// Local time zones:
		"2006-01-02",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		// Explicit time zones:
		"2006-01-02T15:04Z07:00",
		"2006-01-02T15:04Z0700",
		"2006-01-02T15:04:05Z07:00", // time.RFC3339: copied for comparison with other formats
		"2006-01-02T15:04:05Z0700",
	} {
		v, err := time.ParseInLocation(format, s, time.Now().Location())
		if err == nil {
			*tv = TimeValue(v)
			return nil
		}
	}

	// If we did not find a match, return a helpful error message:
	return errors.New(`expected timestamp formatted as "2021-04-16" or "2021-04-16T01:30:59"`)
}

func (tv *TimeValue) Type() string {
	return "time"
}

func (tv *TimeValue) String() string {
	if tv == nil {
		return ""
	}
	t := time.Time(*tv)
	if t.IsZero() {
		return ""
	}

	return t.Format(time.RFC3339)
}
