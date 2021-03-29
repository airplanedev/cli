package root

import (
	"errors"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/login"
	"github.com/airplanedev/cli/pkg/cmd/runs"
	"github.com/airplanedev/cli/pkg/cmd/tasks"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/print"
	isatty "github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// New returns a new root cobra command.
func New() *cobra.Command {
	var output string
	var cfg = &cli.Config{
		Client: &api.Client{},
	}

	cmd := &cobra.Command{
		Use:   "airplane <command>",
		Short: "Airplane CLI",
		Example: heredoc.Doc(`
		airplane tasks deploy -f ./task.yml
		airplane tasks execute my_task

		airplane tasks deploy -f github.com/airplanedev/examples/node/hello-world-javascript/airplane.yml
		airplane tasks execute hello_world
		`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if c, err := conf.ReadDefault(); err == nil {
				cfg.Client.Token = c.Token
			}

			switch output {
			case "json":
				print.DefaultFormatter = print.NewJSONFormatter()
			case "yaml":
				print.DefaultFormatter = print.YAML{}
			case "table":
				print.DefaultFormatter = print.Table{}
			default:
				return errors.New("--output must be (json|yaml|table)")
			}

			return nil
		},
	}

	// Silence usage and errors.
	//
	// Allows us to control how the output looks like.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	// Set usage, help functions.
	cmd.SetUsageFunc(usage)
	cmd.SetHelpFunc(help)

	// Persistent flags, set globally to all commands.
	cmd.PersistentFlags().StringVarP(&cfg.Client.Host, "host", "", api.Host, "Airplane API Host.")
	defaultFormat := "table"
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		defaultFormat = "json"
	}
	cmd.PersistentFlags().StringVarP(&output, "output", "o", defaultFormat, "The format to use for output (json|yaml|table).")

	// Sub-commands.
	cmd.AddCommand(login.New(cfg))
	cmd.AddCommand(tasks.New(cfg))
	cmd.AddCommand(runs.New(cfg))

	return cmd
}
