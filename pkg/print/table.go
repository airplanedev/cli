package print

import (
	"fmt"
	"os"

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
