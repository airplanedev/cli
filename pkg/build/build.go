package build

import (
	"context"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
)

// Request represents a build request.
type Request struct {
	Local   bool
	Client  *api.Client
	Root    string
	Def     definitions.Definition
	TaskID  string
	TaskEnv api.TaskEnv
	Shim    bool
}

// Response represents a build response.
type Response struct {
	ImageURL string
}

// Run runs the build and returns an image URL.
func Run(ctx context.Context, req Request) (*Response, error) {
	if req.Local {
		return local(ctx, req)
	}
	return remote(ctx, req)
}
