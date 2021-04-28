package print

import (
	"github.com/airplanedev/cli/pkg/api"
)

// This struct mirrors api.Task, but with different json/yaml tags.
type printTask struct {
	ID               string               `json:"taskID" yaml:"id"`
	Name             string               `json:"name" yaml:"name"`
	Slug             string               `json:"slug" yaml:"slug"`
	Description      string               `json:"description" yaml:"description"`
	Image            string               `json:"image" yaml:"image"`
	Command          []string             `json:"command" yaml:"command"`
	Arguments        []string             `json:"arguments" yaml:"arguments"`
	Parameters       api.Parameters       `json:"parameters" yaml:"parameters"`
	Constraints      api.RunConstraints   `json:"constraints" yaml:"constraints"`
	Env              api.TaskEnv          `json:"env" yaml:"env"`
	ResourceRequests api.ResourceRequests `json:"resourceRequests" yaml:"resourceRequests"`
	Kind             string               `json:"builder" yaml:"builder"`
	KindOptions      api.KindOptions      `json:"builderConfig" yaml:"builderConfig"`
	Repo             string               `json:"repo" yaml:"repo"`
	Timeout          int                  `json:"timeout" yaml:"timeout"`
}

func printTasks(tasks []api.Task) []printTask {
	pts := make([]printTask, len(tasks))
	for i, t := range tasks {
		pts[i] = printTask(t)
	}
	return pts
}
