package root

import (
	"errors"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/cmd/apikeys"
	"github.com/airplanedev/cli/pkg/cmd/auth"
	"github.com/airplanedev/cli/pkg/cmd/auth/login"
	"github.com/airplanedev/cli/pkg/cmd/auth/logout"
	"github.com/airplanedev/cli/pkg/cmd/configs"
	"github.com/airplanedev/cli/pkg/cmd/runs"
	"github.com/airplanedev/cli/pkg/cmd/tasks"
	"github.com/airplanedev/cli/pkg/cmd/tasks/deploy"
	"github.com/airplanedev/cli/pkg/cmd/tasks/execute"
	"github.com/airplanedev/cli/pkg/cmd/tasks/initcmd"
	"github.com/airplanedev/cli/pkg/cmd/version"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/print"
	"github.com/airplanedev/cli/pkg/trap"
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
		airplane deploy -f ./task.yml
		airplane execute my_task

		airplane deploy -f github.com/airplanedev/examples/node/hello-world-javascript/airplane.yml
		airplane execute hello_world
		`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if c, err := conf.ReadDefault(); err == nil {
				cfg.Client.Token = c.Tokens[cfg.Client.Host]
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

			logger.EnableDebug = cfg.DebugMode
			trap.Printf = logger.Log

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
	cmd.SetVersionTemplate(version.Version() + "\n")

	// Persistent flags, set globally to all commands.
	cmd.PersistentFlags().StringVarP(&cfg.Client.Host, "host", "", api.Host, "Airplane API Host.")
	defaultFormat := "table"
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		defaultFormat = "json"
	}
	cmd.PersistentFlags().StringVarP(&output, "output", "o", defaultFormat, "The format to use for output (json|yaml|table).")
	cmd.PersistentFlags().BoolVar(&cfg.DebugMode, "debug", false, "Whether to produce debugging output.")
	cmd.PersistentFlags().BoolVarP(&cfg.Version, "version", "v", false, "Print the CLI version.")

	// Aliases for popular namespaced commands:
	cmd.AddCommand(initcmd.New(cfg))
	cmd.AddCommand(deploy.New(cfg))
	cmd.AddCommand(execute.New(cfg))
	cmd.AddCommand(login.New(cfg))
	cmd.AddCommand(logout.New(cfg))

	// Sub-commands:
	cmd.AddCommand(apikeys.New(cfg))
	cmd.AddCommand(auth.New(cfg))
	cmd.AddCommand(configs.New(cfg))
	cmd.AddCommand(tasks.New(cfg))
	cmd.AddCommand(runs.New(cfg))
	cmd.AddCommand(version.New(cfg))

	return cmd
}
