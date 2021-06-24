package ignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/airplanedev/cli/pkg/logger"
	"github.com/pkg/errors"
	gitignore "github.com/sabhiram/go-gitignore"
)

// Returns an IgnoreFunc that can be used with airplanedev/archiver to filter
// out files that match a default list or user-provided .airplaneignore.
func GetIgnoreFunc(taskRootPath string) (func(filePath string, info os.FileInfo) (bool, error), error) {
	excludes, err := getIgnorePatterns(taskRootPath)
	if err != nil {
		return nil, err
	}
	logger.Debug("Excludes:\n  %s", strings.Join(excludes, "\n  "))

	ig := gitignore.CompileIgnoreLines(excludes...)
	hasInclusion := false
	for _, pat := range excludes {
		if strings.HasPrefix(pat, "!") {
			hasInclusion = true
			break
		}
	}

	return func(filePath string, info os.FileInfo) (bool, error) {
		// Ignore symbolic links. For example, in Node projects you occasionally see
		// symbolic links to binaries like `.bin/foobar`  which don't exist.
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return false, nil
		}

		relFilePath, err := filepath.Rel(taskRootPath, filePath)
		if err != nil {
			return false, errors.Wrap(err, "getting archive relative path")
		}

		skip := ig.MatchesPath(relFilePath)

		// If we want to skip this file, and it's a directory, then we can skip it only if there
		// are no inclusions.
		// Ideally, we are smarter about this and skip the dir if there's no way for the inclusion
		// pattern to match the dir. However, this is tricky with things like "!foo/" matching
		// any directory named foo/ inside this directory.
		if info.IsDir() && skip && !hasInclusion {
			return false, nil
		}

		if !skip {
			logger.Debug("Including in build archive: %s", relFilePath)
		}
		return !skip, nil
	}, nil
}

func getIgnorePatterns(path string) ([]string, error) {
	// Start with default set of excludes.
	// We exclude the same files regardless of kind because you might have both JS and PY tasks and
	// want pyc files excluded just the same.
	// For inspiration, see:
	// https://github.com/github/gitignore
	// https://github.com/github/gitignore/blob/master/Go.gitignore
	// https://github.com/github/gitignore/blob/master/Node.gitignore
	excludes := []string{
		"*.env",
		"*.pyc",
		".git",
		".gitmodules",
		".hg",
		".idea",
		".next",
		".now",
		".npm",
		".svn",
		".terraform",
		".venv",
		".yarn",
		"__pycache__",
		"bin",
		"dist",
		"node_modules",
		"npm-debug.log",
		"out",
		// Local build artifacts created by `airplane dev`.
		".airplane",
	}

	// Allow user-specified ignore file. Note that users can re-INCLUDE files using !, so if our
	// default excludes skip something necessary they can always add it back.
	const ignorefile = ".airplaneignore"
	bs, err := ioutil.ReadFile(filepath.Join(path, ignorefile))
	switch {
	case os.IsNotExist(err):
		// Nothing additional to append
		return excludes, nil
	case err != nil:
		return nil, err
	}
	fileExcludes := []string{}
	for _, ex := range strings.Split(string(bs), "\n") {
		if ex != "" {
			fileExcludes = append(fileExcludes, ex)
		}
	}
	logger.Debug("Found %s - using %d exclude rule(s):\n  %s", ignorefile, len(fileExcludes), strings.Join(fileExcludes, "\n  "))
	excludes = append(excludes, fileExcludes...)
	return excludes, nil
}
