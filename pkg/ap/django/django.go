package django

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/airplanedev/cli/pkg/ap"
	"github.com/airplanedev/cli/pkg/fsx"
	"github.com/pkg/errors"
)

// Init registers django.
func init() {
	ap.RegisterFramework("django", Open)
}

// Framework implementation.
type Framework struct {
	root string
}

// Open attempts to open at the given path.
func Open(root string) (ap.Framework, error) {
	var managepy = filepath.Join(root, "manage.py")

	if !fsx.Exists(managepy) {
		return nil, fmt.Errorf("django: cannot find manage.py")
	}

	return &Framework{root: root}, nil
}

// ListCommands implementation.
func (f *Framework) ListCommands() ([]string, error) {
	var bin = filepath.Join(f.root, "manage.py")

	cmd := exec.Command(bin, "help", "--commands")
	buf, err := cmd.Output()
	if err != nil {
		var xerr *exec.ExitError
		if errors.As(err, &xerr) {
			return nil, errors.Wrapf(err,
				"exit status=%d stderr=%s",
				xerr.ExitCode(),
				xerr.Stderr,
			)
		}
		return nil, errors.Wrapf(err, "run: %s help --commands", bin)
	}

	var (
		cmds []string
		r    = bytes.NewReader(buf)
		s    = bufio.NewScanner(r)
	)

	for s.Scan() {
		if t := strings.TrimSpace(s.Text()); t != "" {
			cmds = append(cmds, t)
		}
	}

	if err := s.Err(); err != nil {
		return nil, errors.Wrapf(err, "scanning manage.py output")
	}

	sort.Strings(cmds)
	return cmds, nil
}
