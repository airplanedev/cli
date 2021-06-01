package api

import "fmt"

// TaskMissingError implements an exaplainable error.
type TaskMissingError struct {
	app  string
	slug string
}

// Error implementation.
func (err TaskMissingError) Error() string {
	return fmt.Sprintf("task with slug %q does not exist", err.slug)
}

// ExplainError implementation.
func (err TaskMissingError) ExplainError() string {
	return fmt.Sprintf(
		"Follow the URL below to create the task:\n%s",
		err.app+"/tasks/new",
	)
}
