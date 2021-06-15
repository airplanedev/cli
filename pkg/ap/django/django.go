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
	var manage = filepath.Join(f.root, "manage.py")

	bin, err := f.python()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(bin, manage, "help", "--commands")
	buf, err := cmd.Output()
	if err != nil {
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

// Python returns python's bin.
//
// We can actually run manage.py directly, but just to make this
// more robust we'll look for python bin and execute the script using it.
func (f *Framework) python() (string, error) {
	var bins = [...]string{"python3", "python"}

	for _, bin := range bins {
		if p, err := exec.LookPath(bin); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("cannot find one of %v in your $PATH", bins)
}
