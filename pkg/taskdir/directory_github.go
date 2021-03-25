package taskdir

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/airplanedev/cli/pkg/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
)

var (
	// gitHubRegex handles `-f ...` file paths that reference GitHub.
	//
	// Specifically, they should specify the organization and repo name
	// followed by a path from the repo root to an airplane.yml file.
	// They can be optionally suffixed by a git ref selector, using
	// `@ref` syntax, where ref can be a branch name (and soon, a tag or commit, too).
	// As of now, refs must be exact matches, not prefix matches.
	//
	// This syntax is inspired by go modules' go get syntax.
	//
	// More info on the regex: https://regex101.com/r/2DXNxz/1
	gitHubRegex = regexp.MustCompile(`^(?:https:\/\/)?github\.com\/([A-Za-z0-9_.\-]+)\/([A-Za-z0-9_.\-]+)\/([\p{L}0-9_.\-\/]+)(@[A-Za-z0-9_.\-]+)?$`)
)

type gitHubFilePath struct {
	Org  string
	Repo string
	Path string
	Ref  string
}

func parseGitHubFilePath(file string) (gitHubFilePath, error) {
	matches := gitHubRegex.FindAllStringSubmatch(file, -1)
	if len(matches) != 1 || len(matches[0]) < 4 || len(matches[0]) > 5 {
		return gitHubFilePath{}, errors.Errorf("invalid github URL (m=%d): expected github.com/ORG/REPO/PATH/TO/FILE[@REF]", len(matches))
	}

	var fp gitHubFilePath
	fp.Org = matches[0][1]
	fp.Repo = matches[0][2]
	fp.Path = matches[0][3]
	if len(matches[0]) == 5 {
		fp.Ref = strings.TrimPrefix(matches[0][4], "@")
	}

	return fp, nil
}

func openGitHubDirectory(file string) (string, io.Closer, error) {
	fp, err := parseGitHubFilePath(file)
	if err != nil {
		return "", nil, err
	}

	tmpDir, err := ioutil.TempDir("", "airplane-*")
	if err != nil {
		return "", nil, errors.Wrap(err, "creating temporary directory")
	}

	// TODO: consider using git 2.19's --filter option
	// to select just the relevant subdirectory. However, this
	// may not work with go-git.
	//
	// See: https://stackoverflow.com/questions/600079/how-do-i-clone-a-subdirectory-only-of-a-git-repository/52269934#52269934
	r, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: fmt.Sprintf("https://github.com/%s/%s.git", fp.Org, fp.Repo),
	})
	if err != nil {
		// TODO: provide better errors for common edge cases, f.e. lacking auth
		// or a typo in the org/repo, or the configured file not existing.
		return "", nil, errors.Wrap(err, "cloning repo")
	}
	if fp.Ref != "" {
		wt, err := r.Worktree()
		if err != nil {
			return "", nil, errors.Wrap(err, "getting working tree")
		}
		if err := wt.Checkout(&git.CheckoutOptions{
			// TODO: add support for commit and tag references, too.
			Branch: plumbing.NewRemoteReferenceName("origin", fp.Ref),
		}); err != nil {
			return "", nil, errors.Wrap(err, "checking out revision")
		}
	}

	return path.Join(tmpDir, fp.Path), utils.CloseFunc(func() error {
		return errors.Wrap(os.RemoveAll(tmpDir), "cleaning up cloned github repo")
	}), nil
}
