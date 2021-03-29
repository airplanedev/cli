// initcmd defines the implementation of the `airplane tasks init` command.
//
// Even though the command is called "init", we can't name the package "init"
// since that conflicts with the Go init function.
package initcmd

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type config struct {
	cli  *cli.Config
	file string
	from string
}

func New(c *cli.Config) *cobra.Command {
	var cfg = config{cli: c}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a task definition",
		Example: heredoc.Doc(`
			$ airplane tasks init
			$ airplane tasks init -f ./airplane.yml
			$ airplane tasks init --from hello_world
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cmd, cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.file, "file", "f", "airplane.yml", "Path to a file to store task definition")
	cmd.Flags().StringVar(&cfg.from, "from", "", "Slug of an existing task to generate from")

	// --from is the only way to use `init` for now. We'll soon add
	// a way to be prompted through the creation of a task def.
	cli.Must(cmd.MarkFlagRequired("from"))

	return cmd
}

func run(ctx context.Context, cmd *cobra.Command, cfg config) error {
	var client = cfg.cli.Client

	res, err := client.GetTask(ctx, cfg.from)
	if err != nil {
		return errors.Wrap(err, "get task")
	}

	dir, err := taskdir.Open(cfg.file)
	if err != nil {
		return errors.Wrap(err, "opening task directory")
	}
	defer dir.Close()

	if err := dir.WriteDefinition(taskdir.Definition{
		Slug:           res.Slug,
		Name:           res.Name,
		Description:    res.Description,
		Image:          res.Image,
		Command:        res.Command,
		Arguments:      res.Arguments,
		Parameters:     res.Parameters,
		Constraints:    res.Constraints,
		Env:            res.Env,
		ResourceLimits: res.ResourceLimits,
		Builder:        res.Builder,
		BuilderConfig:  res.BuilderConfig,
		Repo:           res.Repo,
		Timeout:        res.Timeout,
	}); err != nil {
		return errors.Wrap(err, "writing task definition")
	}

	cmd.Printf("Created an Airplane task definition for %s in %s\n", res.Name, cfg.file)

	return nil
}
