package print

import (
	"encoding/json"
	"strconv"

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

		ok, jsonObjects := parseArrayOfJsonObject(values)
		if ok {
			printJsonObjects(jsonObjects)
		} else {
			printOutputArray(values)
		}

		if i < len(outputs)-1 {
			fmt.Fprintln(os.Stdout, "")
		}
		i++
	}
}

func parseArrayOfJsonObject(values []json.RawMessage) (bool, []JsonObject) {
	jsonObjects := make([]JsonObject, len(values))
	for i, value := range values {
		if err := json.Unmarshal(value, &jsonObjects[i]); err != nil {
			return false, nil
		}
	}
	return true, jsonObjects
}

func printJsonObjects(objects []JsonObject) {
	keyMap := make(map[string]bool)
	var keyList []string
	for _, object := range objects {
		for key := range object {
			// add key to keyList if not already there
			if _, ok := keyMap[key]; !ok {
				keyList = append(keyList, key)
			}
			keyMap[key] = true
		}
	}

	tw := newTableWriter()
	tw.SetHeader(keyList)
	for _, object := range objects {
		values := make([]string, len(keyList))
		for i, key := range keyList {
			values[i] = object[key]
		}
		tw.Append(values)
	}
	tw.Render()
}

func printOutputArray(values []json.RawMessage) {
	tw := newTableWriter()
	for _, value := range values {
		tw.Append([]string{getCellValue(value)})
	}
	tw.Render()
}

func newTableWriter() *tablewriter.Table {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetBorder(true)
	tw.SetAutoWrapText(false)
	return tw
}

func getCellValue(value json.RawMessage) string {
	var v interface{}
	if err := json.Unmarshal(value, &v); err != nil {
		return string(value)
	}

	switch t := v.(type) {
	case int:
		return strconv.Itoa(t)
	case float32:
	case float64:
		return fmt.Sprintf("%v", t)
	case string:
		return fmt.Sprintf("%s", t)
	default:
		return string(value)
	}
	return ""
}
