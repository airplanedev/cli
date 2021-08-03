package deploy

import (
	"context"
	"io/ioutil"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/build"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// TODO: rename this?
func deployFromTaskYAML(ctx context.Context, cfg config) error {
	client := cfg.client

	// Read in YAML JSON etc.
	buf, err := ioutil.ReadFile(cfg.file)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}
	// TODO: handle JSON format
	// TODO: handle validation, like definitions.UnmarshalDefinition
	var bc api.BuildConfig
	if err := yaml.Unmarshal(buf, &bc); err != nil {
		return errors.Wrap(err, "unmarshalling build definition")
	}

	logger.Debug("build config: %#v", bc)

	task, err := client.GetTask(ctx, bc.Slug)
	if err != nil {
		return err
	}

	resp, err := build.Run(ctx, build.Request{
		Local:       cfg.local,
		Client:      client,
		TaskID:      task.ID,
		BuildConfig: bc,
		TaskEnv:     nil, // TODO
	})
	if err != nil {
		return err
	}
	logger.Debug("build ID: %v", resp.BuildID)
	return nil
}
