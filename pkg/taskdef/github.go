package taskdef

import (
	"io"
	"regexp"
)

type github struct {
	path string
}

var (
	// gitHubRegex handles `-f ...` file paths that reference GitHub.
	//
	// Specifically, they should specify the organization and repo name
	// followed by a path from the repo root to an airplane.yml file.
	// They can be optionally suffixed by a git ref selector, using
	// `@ref` syntax, where ref can be a branch name, tag or commit.
	// As of now, refs must be exact matches, not prefix matches.
	//
	// This syntax is inspired by go modules' go get syntax.
	gitHubRegex = regexp.MustCompile(`^(https://)?(github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)$`)
)

var _ io.Reader = github{}

func (this github) Read(p []byte) (n int, err error) {
	return 0, nil
}
