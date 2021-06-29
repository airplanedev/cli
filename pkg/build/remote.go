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
	"github.com/airplanedev/cli/pkg/build/ignore"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir/definitions"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

func remote(ctx context.Context, req Request) (*Response, error) {
	buildLog(api.LogLevelInfo, logger.Gray("Building with %s as root...", relpath(req.Root)))

	// Before performing a remote build, we must first update kind/kindOptions
	// since the remote build relies on pulling those from the tasks table (for now).
	if err := updateKindAndOptions(ctx, req.Client, req.Def, req.Shim); err != nil {
		return nil, err
	}

	buildLog(api.LogLevelInfo, logger.Gray("Authenticating with Airplane..."))
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

	uploadID, err := uploadArchive(ctx, req.Root, req.Client, archivePath)
	if err != nil {
		return nil, err
	}

	build, err := req.Client.CreateBuild(ctx, api.CreateBuildRequest{
		TaskID:         req.TaskID,
		SourceUploadID: uploadID,
		Env:            req.TaskEnv,
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

func updateKindAndOptions(ctx context.Context, client *api.Client, def definitions.Definition, shim bool) error {
	task, err := client.GetTask(ctx, def.Slug)
	if err != nil {
		return err
	}

	kind, kindOptions, err := def.GetKindAndOptions()
	if err != nil {
		return err
	}
	// Conditionally instruct the remote builder API to perform a shim-based build.
	if shim {
		kindOptions["shim"] = "true"
	}

	_, err = client.UpdateTask(ctx, api.UpdateTaskRequest{
		Kind:        kind,
		KindOptions: kindOptions,

		// The following fields are not updated until after the build finishes.
		Slug:             task.Slug,
		Name:             task.Name,
		Description:      task.Description,
		Image:            task.Image,
		Command:          task.Command,
		Arguments:        task.Arguments,
		Parameters:       task.Parameters,
		Constraints:      task.Constraints,
		Env:              task.Env,
		ResourceRequests: task.ResourceRequests,
		Resources:        task.Resources,
		Repo:             task.Repo,
		Timeout:          task.Timeout,
	})
	if err != nil {
		return errors.Wrapf(err, "updating task %s", def.Slug)
	}

	return nil
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

	var err error
	arch := archiver.NewTarGz()
	arch.Tar.IncludeFunc, err = ignore.Func(root)
	if err != nil {
		return err
	}

	if err := arch.Archive(sources, archivePath); err != nil {
		return errors.Wrap(err, "building archive")
	}

	return nil
}

func uploadArchive(ctx context.Context, root string, client *api.Client, archivePath string) (string, error) {
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

	buildLog(api.LogLevelInfo, logger.Gray("Uploading %s build archive...",
		humanize.Bytes(uint64(sizeBytes)),
	))

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

// Relpath returns the relative using root and the cwd.
func relpath(root string) string {
	if path, err := os.Getwd(); err == nil {
		if rp, err := filepath.Rel(path, root); err == nil {
			if len(rp) == 0 || rp == "." {
				// "." can be missed easily, change it to ./
				return "./"
			}
			return "./" + rp
		}
	}
	return root
}
