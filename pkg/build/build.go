package build

import (
	"context"
	"fmt"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
)

// Request represents a build request.
type Request struct {
	Builder BuilderKind
	Client  *api.Client
	Dir     taskdir.TaskDirectory
	Def     definitions.Definition
	TaskID  string
	TaskEnv api.TaskEnv
}

// Response represents a build response.
type Response struct {
	ImageURL string
}

// Run runs the build and returns an image URL.
func Run(ctx context.Context, req Request) (*Response, error) {
	switch req.Builder {
	case BuilderKindLocal:
		return local(ctx, req)
	case BuilderKindRemote:
		return remote(ctx, req)
	default:
		return nil, fmt.Errorf("build: unknown builder %q", req.Builder)
	}
}
