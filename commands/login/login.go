package login

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/airplanedev/cli/pkg/cli"
	"github.com/airplanedev/cli/pkg/conf"
	"github.com/airplanedev/cli/pkg/token"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// New returns a new login command.
func New(c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Airplane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), c)
		},
	}
	return cmd
}

// Run runs the login command.
func run(ctx context.Context, c *cli.Config) error {
	cfg, err := conf.ReadDefault()

	if errors.Is(err, conf.ErrMissing) {
		srv, err := token.NewServer(ctx)
		if err != nil {
			return err
		}
		defer srv.Close()

		open(loginURL(c.Client.Host, srv.URL()))

		select {
		case <-ctx.Done():
			return ctx.Err()

		case token := <-srv.Token():
			cfg.Token = token
		}

		if err := conf.WriteDefault(cfg); err != nil {
			return err
		}
	}

	fmt.Printf("You're all set!\n\nTo see what tasks you can run, try `$ airplane list`\n")
	return nil
}

// LoginURL returns the CLI login URL.
func loginURL(host, redirect string) string {
	uri := &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/i/cli/getToken",
		RawQuery: url.Values{
			"redirect": []string{redirect},
		}.Encode(),
	}
	return uri.String()
}

// Open attempts to open the URL in the browser.
//
// It uses `open(1)` on darwin and simply prints the URL
// on other operating systems.
func open(url string) {
	if runtime.GOOS == "darwin" {
		if err := exec.Command("open", url).Run(); err == nil {
			return
		}
	}

	fmt.Printf("Visit %s to complete logging in\n", url)
}
