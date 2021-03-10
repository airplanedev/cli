package browser

import (
	"os/exec"
	"strings"
)

// Open attempts to open the browser for `os` at `url`.
func Open(os, url string) error {
	var cmd string
	var args []string

	switch os {
	case "darwin":
		cmd = "open"
		args = append(args, url)

	case "windows":
		cmd = "cmd"
		r := strings.NewReplacer("&", "^&")
		args = append(args, "/c", "start", r.Replace(url))

	default:
		cmd = "xdg-open"
		args = append(args, url)
	}

	bin, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}

	return exec.Command(bin, args...).Run()
}
