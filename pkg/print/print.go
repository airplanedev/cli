package print

import (
	"github.com/airplanedev/cli/pkg/api"
)

var (
	// DefaultFormatter is the default formatter to use.
	//
	// It defaults to the `table` formatter which prints
	// to the CLI using the tablewriter package.
	DefaultFormatter Formatter = Table{}
)

// Formatter represents an output formatter.
type Formatter interface {
	tasks([]api.Task)
	task(api.Task)
	runs([]api.Run)
	run(api.Run)
	outputs(api.Outputs)
}

// Tasks prints the given slice of tasks using the default formatter.
func Tasks(tasks []api.Task) {
	DefaultFormatter.tasks(tasks)
}

// Task prints a single task.
func Task(task api.Task) {
	DefaultFormatter.task(task)
}

// Runs prints the given runs.
func Runs(runs []api.Run) {
	DefaultFormatter.runs(runs)
}

// Run prints a single run.
func Run(run api.Run) {
	DefaultFormatter.run(run)
}

func Outputs(outputs api.Outputs) {
	DefaultFormatter.outputs(outputs)
}
