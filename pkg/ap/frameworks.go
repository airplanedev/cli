package ap

import (
	"fmt"
	"sort"
)

var (
	// Frameworks is a mapping of framework names to framework adapters.
	frameworks = make(map[string]FrameworkAdapter)
)

// FrameworkAdapter accepts a root path and returns
// a new initialized framework.
//
// Typically, the method checks to see if the root path
// has the necessary files for the framework, if one of the
// files is missing the adapter returns an error.
type FrameworkAdapter func(root string) (Framework, error)

// RegisterFramework registers a framework with name and adapter.
//
// If the framework is already registered the function panics.
func RegisterFramework(name string, adapter FrameworkAdapter) {
	if _, ok := frameworks[name]; ok {
		panic(fmt.Sprintf("ap: %q framework is already registered", name))
	}
	frameworks[name] = adapter
}

// LookupFramework returns a new initialized framework with
// name and root path.
//
// If the framework does not exist, an error is returned.
func LookupFramework(name, root string) (Framework, error) {
	if adapter, ok := frameworks[name]; ok {
		return adapter(root)
	}
	return nil, fmt.Errorf("ap: framework %q does not exist", name)
}

// ListFrameworks returns a slice of sorted framework names.
func ListFrameworks() (names []string) {
	for name := range frameworks {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}

// Framework represents a framework.
type Framework interface {
	// ListCommands lists all commands.
	//
	// When there are no commands, a nil slice and error are returned.
	ListCommands() ([]string, error)
}
