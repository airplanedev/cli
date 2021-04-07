package build

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/mholt/archiver/v3"
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

	// TODO: filter out files/directories that match .dockerignore
	if err := archiver.Archive(sources, archivePath); err != nil {
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
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
