package build

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

// Args represent build arguments.
//
// The arguments depend on the builder used.
//
// TODO(amir): refine this, we need the build args
// to be typed so we can show their usage in the CLI.
type Args map[string]string

// RegistryAuth represents the registry auth.
type RegistryAuth struct {
	Token string
	Repo  string
}

// Host returns the registry hostname.
func (r RegistryAuth) host() string {
	return strings.SplitN(r.Repo, "/", 2)[0]
}

// Config configures a builder.
type Config struct {
	// Root is the root directory.
	//
	// It must be an absolute path to the project directory.
	Root string

	// Builder is the builder name to use.
	//
	// There are various built-in builders, along with the docker
	// builder and manual builder.
	//
	// If empty, it assumes a manual builder.
	Builder string

	// Args are the build arguments to use.
	//
	// When nil, it uses an empty map of arguments.
	Args Args

	// Auth represents the registry auth to use.
	//
	// If nil, New returns an error.
	Auth *RegistryAuth

	// Writer is the writer to output docker build
	// status to.
	//
	// TODO(amir): we may want to read the output stream
	// detect an error and return it.
	//
	// If empty, os.Stderr is used.
	Writer io.Writer
}

// Builder implements an image builder.
type Builder struct {
	root   string
	name   string
	args   Args
	writer io.Writer
	auth   *RegistryAuth
	client *client.Client
}

// New returns a new builder with c.
func New(c Config) (*Builder, error) {
	if !filepath.IsAbs(c.Root) {
		return nil, fmt.Errorf("build: expected an absolute path, got %q", c.Root)
	}

	if c.Builder == "" {
		c.Builder = "manual"
	}

	if c.Args == nil {
		c.Args = make(Args)
	}

	if c.Writer == nil {
		c.Writer = os.Stderr
	}

	if c.Auth == nil {
		return nil, fmt.Errorf("build: builder requires registry auth")
	}

	client, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	return &Builder{
		root:   c.Root,
		name:   c.Builder,
		args:   c.Args,
		writer: c.Writer,
		auth:   c.Auth,
		client: client,
	}, nil
}

// Build runs the docker build.
//
// Depending on the configured `Config.Builder` the method verifies that
// the directory can be built and all necessary files exist.
//
// The method creates a Dockerfile depending on the configured builder
// and adds it to the tree, it passes the tree as the build context
// and initializes the build.
func (b *Builder) Build(ctx context.Context, taskID, version string) (*types.ImageSummary, error) {
	var repo = b.auth.Repo
	var name = "task-" + sanitizeTaskID(taskID)
	var tag = repo + "/" + name + ":" + version

	tree, err := NewTree()
	if err != nil {
		return nil, errors.Wrap(err, "new tree")
	}
	defer tree.Close()

	buf, err := b.dockerfile()
	if err != nil {
		return nil, errors.Wrap(err, "creating dockerfile")
	}

	if err := tree.Write("Dockerfile", strings.NewReader(buf)); err != nil {
		return nil, errors.Wrap(err, "writing dockerfile")
	}

	if err := tree.Copy(b.root); err != nil {
		return nil, errors.Wrapf(err, "copy %q", b.root)
	}

	bc, err := tree.Archive()
	if err != nil {
		return nil, errors.Wrap(err, "archive tree")
	}
	defer bc.Close()

	opts := types.ImageBuildOptions{
		Tags:        []string{tag},
		BuildArgs:   map[string]*string{},
		Platform:    "linux/amd64",
		AuthConfigs: b.authconfigs(),
	}

	resp, err := b.client.ImageBuild(ctx, bc, opts)
	if err != nil {
		return nil, errors.Wrap(err, "image build")
	}
	defer resp.Body.Close()

	// TODO(amir): read and abort on any build errors, including the surrounding
	// lines.
	if _, err := io.Copy(b.writer, resp.Body); err != nil {
		return nil, errors.Wrap(err, "copy output")
	}

	images, err := b.client.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "image list")
	}

	for _, img := range images {
		for _, t := range img.RepoTags {
			if t == tag {
				return &img, nil
			}
		}
	}

	return nil, fmt.Errorf("build: image with the tag %q was not found", tag)
}

// Push pushes the given image.
func (b *Builder) Push(ctx context.Context, tag string) error {
	authjson, err := json.Marshal(b.registryAuth())
	if err != nil {
		return err
	}

	resp, err := b.client.ImagePush(ctx, tag, types.ImagePushOptions{
		RegistryAuth: base64.URLEncoding.EncodeToString(authjson),
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	// TODO(amir): read and abort on any errors.
	if _, err := io.Copy(b.writer, resp); err != nil {
		return errors.Wrap(err, "image push")
	}

	return nil
}

// Dockerfile returns the dockerfile for the builder.
func (b *Builder) dockerfile() (string, error) {
	if b.args["entrypoint"] == "" {
		return "", fmt.Errorf("build: .entrypoint is required")
	}

	switch b.name {
	case "go":
		return golang(b.root, b.args)
	case "deno":
		return deno(b.root, b.args)
	case "python":
		return python(b.root, b.args)
	case "node":
		return node(b.root, b.args)
	default:
		return "", errors.Errorf("build: unknown builder type %q", b.name)
	}
}

// RegistryAuth returns the registry auth.
func (b *Builder) registryAuth() types.AuthConfig {
	return types.AuthConfig{
		Username: "oauth2accesstoken",
		Password: b.auth.Token,
	}
}

// Authconfigs returns the authconfigs to use.
func (b *Builder) authconfigs() map[string]types.AuthConfig {
	return map[string]types.AuthConfig{
		b.auth.host(): b.registryAuth(),
	}
}

// SanitizeTaskID sanitizes the given task ID.
//
// Names may only contain lowercase letters, numbers, and
// hyphens, and must begin with a letter and end with a letter or number.
//
// We are planning to tweak our team/task ID generation to fit this:
// https://linear.app/airplane/issue/AIR-355/restrict-task-id-charset
//
// The following string manipulations won't matter for non-ksuid
// IDs (the current scheme).
func sanitizeTaskID(s string) string {
	s = strings.ToLower(s)
	if unicode.IsDigit(rune(s[len(s)-1])) {
		s = s[:len(s)-1] + "a"
	}
	return s
}
