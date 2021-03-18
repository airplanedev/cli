package print

import (
	"fmt"
	"os"
	"time"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/olekukonko/tablewriter"
)

// Table implements a table formatter.
//
// Its zero-value is ready for use.
type Table struct{}

// Tasks implementation.
func (t Table) tasks(tasks []api.Task) {
	tw := tablewriter.NewWriter(os.Stderr)
	tw.SetBorder(false)
	tw.SetHeader([]string{"name", "slug", "builder", "arguments"})

	for _, t := range tasks {
		var builder = t.Builder

		if builder == "" {
			builder = "manual"
		}

		tw.Append([]string{
			t.Name,
			t.Slug,
			t.Builder,
			fmt.Sprintf("%v", t.Arguments),
		})
	}

	tw.Render()
}

// Task implementation.
func (t Table) task(task api.Task) {
	t.tasks([]api.Task{task})
}

// Runs implementation.
func (t Table) runs(runs []api.Run) {
	tw := tablewriter.NewWriter(os.Stderr)
	tw.SetBorder(false)
	tw.SetHeader([]string{"id", "status", "created at", "ended at"})

	for _, run := range runs {
		var endedAt string

		switch {
		case run.SucceededAt != nil:
			endedAt = run.SucceededAt.Format(time.RFC3339)
		case run.FailedAt != nil:
			endedAt = run.FailedAt.Format(time.RFC3339)
		case run.CancelledAt != nil:
			endedAt = run.CancelledAt.Format(time.RFC3339)
		}

		tw.Append([]string{
			run.RunID,
			fmt.Sprintf("%s", run.Status),
			run.CreatedAt.Format(time.RFC3339),
			endedAt,
		})
	}

	tw.Render()
}

// Run implementation.
func (t Table) run(run api.Run) {
	t.runs([]api.Run{run})
}
