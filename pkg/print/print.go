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
	apiKeys([]api.APIKey)
	tasks([]api.Task)
	task(api.Task)
	runs([]api.Run)
	run(api.Run)
	outputs(api.Outputs)
	config(api.Config)
}

// APIKeys prints one or more API keys.
func APIKeys(apiKeys []api.APIKey) {
	DefaultFormatter.apiKeys(apiKeys)
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

// Outputs prints a collection of outputs.
func Outputs(outputs api.Outputs) {
	DefaultFormatter.outputs(outputs)
}

// Config prints a single config var.
func Config(config api.Config) {
	DefaultFormatter.config(config)
}
