package build

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/airplanedev/archiver"
	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	dockerFileUtils "github.com/docker/docker/pkg/fileutils"
	"github.com/dustin/go-humanize"
	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/pkg/errors"
)

func remote(ctx context.Context, req Request) (*Response, error) {
	registry, err := req.Client.GetRegistryToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting registry token")
	}

	tmpdir, err := ioutil.TempDir("", "airplane-builds-")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary directory for remote build")
	}
	defer os.RemoveAll(tmpdir)

	archivePath := path.Join(tmpdir, "archive.tar.gz")
	if err := archiveTaskDir(req.Def, req.Root, archivePath); err != nil {
		return nil, err
	}

	uploadID, err := uploadArchive(ctx, req.Client, archivePath)
	if err != nil {
		return nil, err
	}

	build, err := req.Client.CreateBuild(ctx, api.CreateBuildRequest{
		TaskID:         req.TaskID,
		SourceUploadID: uploadID,
		Env:            req.TaskEnv,
		Shim:           req.Shim,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating build")
	}
	logger.Debug("Created build with id=%s", build.Build.ID)

	if err := waitForBuild(ctx, req.Client, build.Build.ID); err != nil {
		return nil, err
	}

	imageURL := fmt.Sprintf("%s/task-%s:%s",
		registry.Repo,
		sanitizeTaskID(req.TaskID),
		build.Build.ID,
	)

	return &Response{
		ImageURL: imageURL,
	}, nil
}

func archiveTaskDir(def definitions.Definition, root string, archivePath string) error {
	// mholt/archiver takes a list of "sources" (files/directories) that will
	// be included in the root of the archive. In our case, we want the root of
	// the archive to be the contents of the task directory, rather than the
	// task directory itself.
	var sources []string
	if files, err := ioutil.ReadDir(root); err != nil {
		return errors.Wrap(err, "inspecting files in task root")
	} else {
		for _, f := range files {
			sources = append(sources, path.Join(root, f.Name()))
		}
	}

	arch := archiver.NewTarGz()

	kind, _, err := def.GetKindAndOptions()
	if err != nil {
		return err
	}

	arch.Tar.IncludeFunc, err = getIgnoreFunc(root, kind)
	if err != nil {
		return err
	}

	if err := arch.Archive(sources, archivePath); err != nil {
		return errors.Wrap(err, "building archive")
	}

	return nil
}

// Returns an IgnoreFunc that can be used with airplanedev/archiver to filter
// out files that match a default (or user-provided) .dockerignore.
//
// This is modeled off of docker/cli.
// See: https://github.com/docker/cli/blob/a32cd16160f1b41c1c4ae7bee4dac929d1484e59/vendor/github.com/docker/docker/pkg/archive/archive.go#L738
func getIgnoreFunc(taskRootPath string, kind api.TaskKind) (func(filePath string, info os.FileInfo) (bool, error), error) {
	excludes, err := getIgnorePatterns(taskRootPath)
	if err != nil {
		return nil, err
	}

	pm, err := dockerFileUtils.NewPatternMatcher(excludes)
	if err != nil {
		return nil, errors.Wrap(err, "parsing dockerignore patterns")
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

		skip, err := pm.Matches(relFilePath)
		if err != nil {
			return false, errors.Wrap(err, "matching file")
		}

		// If we want to skip this file and it's a directory
		// then we should first check to see if there's an
		// inclusion pattern (e.g. !dir/file) that starts with this
		// dir. If so then we can't skip this dir.
		if info.IsDir() && skip {
			for _, pat := range pm.Patterns() {
				if !pat.Exclusion() {
					continue
				}
				if strings.HasPrefix(pat.String()+string(filepath.Separator), relFilePath+string(filepath.Separator)) {
					// There is a pattern in this directory that should be included, so
					// we can't skip this directory.
					logger.Debug("Including in build archive: %s", relFilePath)
					return true, nil
				}
			}

			return false, nil
		}

		if !skip {
			logger.Debug("Including in build archive: %s", relFilePath)
		}
		return !skip, nil
	}, nil
}

// readIgnorefile reads a .dockerignore-like file
// Based off of github.com/docker/cli@v20.10.6/cli/command/image/build/dockerignore.go
// Reference: https://docs.docker.com/engine/reference/builder/#dockerignore-file
func readIgnorefile(contextDir, filename string) ([]string, error) {
	var excludes []string

	f, err := os.Open(filepath.Join(contextDir, filename))
	switch {
	case os.IsNotExist(err):
		return excludes, nil
	case err != nil:
		return nil, err
	}
	defer f.Close()

	return dockerignore.ReadAll(f)
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
		"**/*.env",
		"**/*.pyc",
		"**/.next",
		"**/.now",
		"**/.npm",
		"**/.venv",
		"**/.yarn",
		"**/__pycache__",
		"**/bin",
		"**/dist",
		"**/node_modules",
		"**/npm-debug.log",
		"**/out",
		".git",
		".gitmodules",
		".hg",
		".svn",
	}

	// Allow user-specified ignore file. Note that users can re-INCLUDE files using !, so if our
	// default excludes skip something necessary they can always add it back.

	// Prefer .airplaneignore over .dockerignore
	for _, ignorefile := range []string{".airplaneignore", ".dockerignore"} {
		ex, err := readIgnorefile(path, ignorefile)
		if err != nil {
			return nil, errors.Wrapf(err, "reading %s", ignorefile)
		}
		if len(ex) > 0 {
			logger.Debug("Found %s - using %d exclude rule(s)", ignorefile, len(ex))
			excludes = append(excludes, ex...)
			return excludes, nil
		}
	}
	return excludes, nil
}

func uploadArchive(ctx context.Context, client *api.Client, archivePath string) (string, error) {
	archive, err := os.OpenFile(archivePath, os.O_RDONLY, 0)
	if err != nil {
		return "", errors.Wrap(err, "opening archive file")
	}
	defer archive.Close()

	info, err := archive.Stat()
	if err != nil {
		return "", errors.Wrap(err, "stat on archive file")
	}
	sizeBytes := int(info.Size())

	buildLog(api.LogLevelInfo, logger.Gray("Uploading %s build archive...", humanize.Bytes(uint64(sizeBytes))))

	upload, err := client.CreateBuildUpload(ctx, api.CreateBuildUploadRequest{
		SizeBytes: sizeBytes,
	})
	if err != nil {
		return "", errors.Wrap(err, "creating upload")
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", upload.WriteOnlyURL, archive)
	if err != nil {
		return "", errors.Wrap(err, "creating GCS upload request")
	}
	req.Header.Add("X-Goog-Content-Length-Range", fmt.Sprintf("0,%d", sizeBytes))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "uploading to GCS")
	}
	defer resp.Body.Close()

	logger.Debug("Upload complete: %s", upload.Upload.URL)

	return upload.Upload.ID, nil
}

func waitForBuild(ctx context.Context, client *api.Client, buildID string) error {
	buildLog(api.LogLevelInfo, logger.Gray("Waiting for builder..."))

	t := time.NewTicker(time.Second)

	var since time.Time
	var logs []api.LogItem
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			r, err := client.GetBuildLogs(ctx, buildID, since)
			if err != nil {
				return errors.Wrap(err, "getting build logs")
			}
			if len(r.Logs) > 0 {
				since = r.Logs[len(r.Logs)-1].Timestamp
			}

			newLogs := api.DedupeLogs(logs, r.Logs)
			for _, l := range newLogs {
				text := l.Text
				if strings.HasPrefix(l.Text, "[builder] ") {
					text = logger.Gray(strings.TrimPrefix(text, "[builder] "))
				}

				buildLog(l.Level, text)
			}
			logs = append(logs, newLogs...)

			b, err := client.GetBuild(ctx, buildID)
			if err != nil {
				return errors.Wrap(err, "getting build")
			}

			if b.Build.Status.Stopped() {
				switch b.Build.Status {
				case api.BuildCancelled:
					logger.Log("\nBuild " + logger.Bold(logger.Yellow("cancelled")))
					return errors.New("Build cancelled")
				case api.BuildFailed:
					logger.Log("\nBuild " + logger.Bold(logger.Red("failed")))
					return errors.New("Build failed")
				case api.BuildSucceeded:
					logger.Log("\nBuild " + logger.Bold(logger.Green("succeeded")))
				}

				return nil
			}
		}
	}
}

func buildLog(level api.LogLevel, msg string, args ...interface{}) {
	if level == api.LogLevelDebug {
		logger.Log("["+logger.Yellow("build")+"] ["+logger.Blue("debug")+"] "+msg, args...)
	} else {
		logger.Log("["+logger.Yellow("build")+"] "+msg, args...)
	}
}
