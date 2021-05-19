package build

import (
	"context"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/configs"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/pkg/errors"
)

func local(ctx context.Context, req Request) (*Response, error) {
	registry, err := req.Client.GetRegistryToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting registry token")
	}

	buildEnv, err := getBuildEnv(ctx, req.Client, req.Def)
	if err != nil {
		return nil, err
	}

	kind, options, err := req.Def.GetKindAndOptions()
	if err != nil {
		return nil, err
	}
	args := make(map[string]string, len(options))
	for k, v := range options {
		if sv, ok := v.(string); !ok {
			return nil, errors.New("unexpected non-string option for builder arg")
		} else {
			args[k] = sv
		}
	}
	b, err := New(LocalConfig{
		Root:    req.Dir.DefinitionRootPath(),
		Builder: string(kind),
		Args:    args,
		Auth: &RegistryAuth{
			Token: registry.Token,
			Repo:  registry.Repo,
		},
		BuildEnv: buildEnv,
	})
	if err != nil {
		return nil, errors.Wrap(err, "new build")
	}

	logger.Log("Building...")
	resp, err := b.Build(ctx, req.TaskID, "latest")
	if err != nil {
		return nil, errors.Wrap(err, "build")
	}

	logger.Log("Pushing...")
	if err := b.Push(ctx, resp.ImageURL); err != nil {
		return nil, errors.Wrap(err, "push")
	}

	return resp, nil
}

// Retreives a build env from def - looks for env vars starting with BUILD_ and either uses the
// string literal or looks up the config value.
func getBuildEnv(ctx context.Context, client *api.Client, def definitions.Definition) (map[string]string, error) {
	buildEnv := make(map[string]string)
	for k, v := range def.Env {
		if v.Value != nil {
			buildEnv[k] = *v.Value
		} else if v.Config != nil {
			nt, err := configs.ParseName(*v.Config)
			if err != nil {
				return nil, err
			}
			res, err := client.GetConfig(ctx, api.GetConfigRequest{
				Name:       nt.Name,
				Tag:        nt.Tag,
				ShowSecret: true,
			})
			if err != nil {
				return nil, err
			}
			buildEnv[k] = res.Config.Value
		}
	}
	return buildEnv, nil
}
