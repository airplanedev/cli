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
	"github.com/airplanedev/cli/pkg/taskdir"
	dockerBuild "github.com/docker/cli/cli/command/image/build"
	dockerFileUtils "github.com/docker/docker/pkg/fileutils"
	"github.com/pkg/errors"
)

func Remote(ctx context.Context, dir taskdir.TaskDirectory, client *api.Client, taskRevisionID string) error {
	tmpdir, err := ioutil.TempDir("", "airplane-builds-")
	if err != nil {
		return errors.Wrap(err, "creating temporary directory for remote build")
	}
	defer os.RemoveAll(tmpdir)

	archivePath := path.Join(tmpdir, "archive.tar.gz")
	if err := archiveTaskDir(dir, archivePath); err != nil {
		return err
	}

	uploadID, err := uploadArchive(ctx, client, archivePath)
	if err != nil {
		return err
	}

	build, err := client.CreateBuild(ctx, api.CreateBuildRequest{
		TaskRevisionID: taskRevisionID,
		SourceUploadID: uploadID,
	})
	if err != nil {
		return errors.Wrap(err, "creating build")
	}

	if err := waitForBuild(ctx, client, build.Build.ID); err != nil {
		return err
	}

	return nil
}

func archiveTaskDir(dir taskdir.TaskDirectory, archivePath string) error {
	// mholt/archiver takes a list of "sources" (files/directories) that will
	// be included in the root of the archive. In our case, we want the root of
	// the archive to be the contents of the task directory, rather than the
	// task directory itself.
	var sources []string
	if files, err := ioutil.ReadDir(dir.DefinitionRootPath()); err != nil {
		return errors.Wrap(err, "inspecting files in task root")
	} else {
		for _, f := range files {
			sources = append(sources, path.Join(dir.DefinitionRootPath(), f.Name()))
		}
	}

	// reference: https://docs.docker.com/engine/reference/builder/#dockerignore-file
	excludes, err := dockerBuild.ReadDockerignore(dir.DefinitionRootPath())
	if err != nil {
		return errors.Wrap(err, "reading .dockerignore")
	}
	if len(excludes) == 0 {
		// If a .dockerignore was not provided, use a default based on the builder.
		def, err := dir.ReadDefinition()
		if err != nil {
			return err
		}
		defaultExcludes := []string{
			".git",
			"*.env",
		}
		// For inspiration, see: https://github.com/github/gitignore
		switch BuilderName(def.Builder) {
		case BuilderNameGo:
			// https://github.com/github/gitignore/blob/master/Go.gitignore
			excludes = append(defaultExcludes, []string{
				"vendor",
			}...)
		case BuilderNameDeno:
			excludes = defaultExcludes
		case BuilderNamePython:
			excludes = defaultExcludes
		case BuilderNameNode:
			// https://github.com/github/gitignore/blob/master/Node.gitignore
			excludes = append(defaultExcludes, []string{
				"node_modules",
				".npm",
				".next",
				"out",
				"dist",
				".yarn",
			}...)
		case BuilderNameDocker:
			excludes = defaultExcludes
		default:
			return errors.Errorf("build: unknown builder type %s", def.Builder)
		}
	}

	a := archiver.NewTarGz()
	pm, err := dockerFileUtils.NewPatternMatcher(excludes)
	if err != nil {
		return errors.Wrap(err, "parsing dockerignore patterns")
	}
	a.Tar.IncludeFunc = func(filePath string, info os.FileInfo) (bool, error) {
		// This is modeled off of docker/cli.
		// See: https://github.com/docker/cli/blob/a32cd16160f1b41c1c4ae7bee4dac929d1484e59/vendor/github.com/docker/docker/pkg/archive/archive.go#L738
		relFilePath, err := filepath.Rel(dir.DefinitionRootPath(), filePath)
		if err != nil {
			return false, errors.Wrap(err, "getting archive relative path")
		}
		logger.Log("checking relative file path: %s", relFilePath)

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
					logger.Debug("Including directory from inclusion rule: %s", relFilePath)
					return true, nil
				}
			}

			return false, nil
		}

		return !skip, nil
	}
	if err := a.Archive(sources, archivePath); err != nil {
		return errors.Wrap(err, "building archive")
	}

	return nil
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

	upload, err := client.CreateBuildUpload(ctx, api.CreateBuildUploadRequest{
		SizeBytes: sizeBytes,
	})
	if err != nil {
		return "", errors.Wrap(err, "creating upload")
	}

	logger.Debug("Uploaded archive to id=%s at url=%s", upload.Upload.ID, upload.Upload.URL)

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

	return upload.Upload.ID, nil
}

func waitForBuild(ctx context.Context, client *api.Client, buildID string) error {
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

			buildPrefix := "[" + logger.Yellow("build") + "] "
			newLogs := api.DedupeLogs(logs, r.Logs)
			for _, l := range newLogs {
				logger.Log(buildPrefix + logger.Gray(l.Text))
			}
			logs = append(logs, newLogs...)

			b, err := client.GetBuild(ctx, buildID)
			if err != nil {
				return errors.Wrap(err, "getting build")
			}

			if b.Build.Status.Stopped() {
				return nil
			}
		}
	}
}
