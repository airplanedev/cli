package build

import (
	"context"
	"os"
	"path"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/airplanedev/cli/pkg/taskdir"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
)

func Remote(ctx context.Context, dir taskdir.TaskDirectory, client *api.Client) error {
	tmpdir := os.TempDir()
	defer os.RemoveAll(tmpdir)

	// Archive the root task directory:
	archiveName := "airplane-build.tar.gz"
	archivePath := path.Join(tmpdir, archiveName)
	// TODO: filter out files/directories that match .dockerignore
	if err := archiver.Archive([]string{dir.DefinitionRootPath()}, archivePath); err != nil {
		return errors.Wrap(err, "building archive")
	}

	req := api.UploadBuildRequest{
		FileName: archiveName,
	}

	// Compute the size of this archive:
	f, err := os.OpenFile(archivePath, os.O_RDONLY, 0)
	if err != nil {
		return errors.Wrap(err, "opening archive file")
	}
	defer f.Close()
	if info, err := f.Stat(); err != nil {
		return errors.Wrap(err, "stat on archive file")
	} else {
		req.SizeBytes = int(info.Size())
	}

	// Upload the archive to Airplane:
	upload, err := client.UploadBuild(ctx, req)
	if err != nil {
		return errors.Wrap(err, "creating upload")
	}
	logger.Debug("Uploaded archive to id=%s at %s", upload.ID, upload.URL)

	// TODO: GCS write to that URL

	// TODO: create the build, referencing this upload
	// TODO: poll the build until it finishes

	// TODO: once this works e2e, we can remove this error:
	return errors.New("remote builds not implemented")
}
