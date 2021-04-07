package build

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
)

func Remote(ctx context.Context, dir taskdir.TaskDirectory, client *api.Client) error {
	tmpdir, err := ioutil.TempDir("", "airplane-builds-")
	if err != nil {
		return errors.Wrap(err, "creating temporary directory for remote build")
	}
	defer os.RemoveAll(tmpdir)

	// Archive the root task directory:
	// TODO: filter out files/directories that match .dockerignore
	archivePath := path.Join(tmpdir, "airplane-build.tar.gz")
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
	if err := archiver.Archive(sources, archivePath); err != nil {
		return errors.Wrap(err, "building archive")
	}

	// Compute the size of this archive:
	var sizeBytes int
	archive, err := os.OpenFile(archivePath, os.O_RDONLY, 0)
	if err != nil {
		return errors.Wrap(err, "opening archive file")
	}
	defer archive.Close()
	if info, err := archive.Stat(); err != nil {
		return errors.Wrap(err, "stat on archive file")
	} else {
		sizeBytes = int(info.Size())
	}

	// Upload the archive to Airplane:
	upload, err := client.CreateBuildUpload(ctx, api.CreateBuildUploadRequest{
		SizeBytes: sizeBytes,
	})
	if err != nil {
		return errors.Wrap(err, "creating upload")
	}

	logger.Debug("Uploaded archive to id=%s at url=%s", upload.Upload.ID, upload.Upload.URL)

	req, err := http.NewRequestWithContext(ctx, "PUT", upload.WriteOnlyURL, archive)
	if err != nil {
		return errors.Wrap(err, "creating GCS upload request")
	}
	req.Header.Add("X-Goog-Content-Length-Range", fmt.Sprintf("0,%d", sizeBytes))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "uploading to GCS")
	}
	defer resp.Body.Close()

	build, err := client.CreateBuild(ctx, api.CreateBuildRequest{
		TaskRevisionID: "todo",
		SourceUploadID: upload.Upload.ID,
	})
	if err != nil {
		return errors.Wrap(err, "creating build")
	}

	b, err := client.GetBuild(ctx, build.Build.ID)
	if err != nil {
		return errors.Wrap(err, "getting build")
	}

	// TODO: wait until done
	if b.Build.Status.IsDone() {
		return nil
	}

	// TODO: once this works e2e, we can remove this error:
	return errors.New("remote builds not implemented")
}
