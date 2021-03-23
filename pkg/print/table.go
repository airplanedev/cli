package print

import (
	"encoding/json"

	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/olekukonko/tablewriter"
)

// Table implements a table formatter.
//
// Its zero-value is ready for use.
type Table struct{}

type JsonObject map[string]string

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

// print outputs as table
func (t Table) outputs(outputs api.Outputs) {
	i := 0
	for key, values := range outputs {

		fmt.Fprintln(os.Stdout, key)

		if isJsonObject(values[0]) {
			printOutputObjects(values)
		} else {
			printOutputArray(values)
		}

		if i < len(outputs)-1 {
			fmt.Fprintln(os.Stdout, "")
		}
		i++
	}
}

func isJsonObject(value json.RawMessage) bool {
	var output JsonObject
	err := json.Unmarshal(value, &output)
	return err == nil
}

func printOutputObjects(values []json.RawMessage) {
	objectArray := make([]JsonObject, len(values))

	keys := make(map[string]bool)

	for i, value := range values {
		var output JsonObject
		if err := json.Unmarshal(value, &output); err != nil {
			fmt.Printf("  Error: %s\n", errors.Cause(err).Error())
		} else {
			objectArray[i] = output

			for key := range output {
				keys[key] = true
			}
		}
	}

	var keyList []string
	for _, object := range objectArray {
		for key := range object {
			keyList = append(keyList, key)
		}
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetBorder(true)
	tw.SetHeader(keyList)

	for _, object := range objectArray {
		values := make([]string, len(keyList))
		for i, key := range keyList {
			values[i] = object[key]
		}
		tw.Append(values)
	}

	tw.Render()
}

func printOutputArray(values []json.RawMessage) {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetBorder(true)

	for _, value := range values {
		tw.Append([]string{string(value)})
	}

	tw.Render()
}
