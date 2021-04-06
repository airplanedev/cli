package build

import (
	"context"
	"io/ioutil"
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
	logger.Debug("tmpdir: %s", tmpdir)
	// defer os.RemoveAll(tmpdir)

	// Archive the root task directory:
	// TODO: filter out files/directories that match .dockerignore
	archiveName := "airplane-build.tar.gz"
	archivePath := path.Join(tmpdir, archiveName)
	// We want to produce an archive where the contents of the archive
	// are the files inside of `dir.DefinitionRootPath()`, rather than
	// a directory containing those files. Therefore, we need to produce
	// a list of files/directories within the root directory instead of
	// directly providing mholt/archiver with `dir.DefinitionRootPath()`.
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
	resp, err := client.UploadBuild(ctx, req)
	if err != nil {
		return errors.Wrap(err, "creating upload")
	}
	logger.Debug("Uploaded archive to id=%s at %s", resp.Upload.ID, resp.Upload.URL)

	// TODO: GCS write to that URL

	// TODO: create the build, referencing this upload
	// TODO: poll the build until it finishes

	// TODO: once this works e2e, we can remove this error:
	return errors.New("remote builds not implemented")
}
