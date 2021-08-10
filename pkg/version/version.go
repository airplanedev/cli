package version

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/airplanedev/cli/pkg/logger"
)

// Set by Go Releaser.
var (
	version string = "<unknown>"
	date    string = "<unknown>"
)

func Get() string {
	return version
}

func Date() string {
	return date
}

const releaseURL = "https://api.github.com/repos/airplanedev/cli/releases?per_page=1"

type release struct {
	Name string `json:"name"`
}

// CheckLatest queries the GitHub API for newer releases and prints a warning if the CLI is outdated.
func CheckLatest(ctx context.Context) error {
	latest, err := getLatest(ctx)
	if err != nil {
		return err
	}
	if latest == "" || version == "<unknown>" {
		// No version found or CLI version unknown - pass silently.
		return nil
	}
	// Assumes not matching latest means you are behind:
	if latest != "v"+version {
		logger.Warning("A newer version of the Airplane CLI is available: %s", latest)
		logger.Suggest(
			"Visit the docs for upgrade instructions:",
			"https://docs.airplane.dev/platform/install-the-airplane-cli#upgrading-the-cli",
		)
	}
	return nil
}

func getLatest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}
	if len(releases) == 0 {
		return "", nil
	}
	return releases[0].Name, nil
}
