package build

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/airplanedev/cli/pkg/api"
	"github.com/airplanedev/cli/pkg/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	dockerJSONMessage "github.com/docker/docker/pkg/jsonmessage"
	"github.com/mattn/go-isatty"
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

// BuilderKind represents where the Docker build should take place.
//
// For BuilderKindLocal, users need to have Docker installed and running.
// For BuilderKindRemote, the build will happen on Airplane servers.
type BuilderKind string

const (
	BuilderKindLocal  BuilderKind = "local"
	BuilderKindRemote BuilderKind = "remote"
)

func ToBuilderKind(s string) (BuilderKind, error) {
	switch s {
	case string(BuilderKindLocal):
		return BuilderKindLocal, nil
	case string(BuilderKindRemote):
		return BuilderKindRemote, nil
	default:
		return BuilderKind(""), errors.Errorf("Unknown builder: %s", s)
	}
}

// Host returns the registry hostname.
func (r RegistryAuth) host() string {
	return strings.SplitN(r.Repo, "/", 2)[0]
}

// LocalConfig configures a (local) builder.
type LocalConfig struct {
	// Kind describes how the build should be performed, such as
	// whether it should use the local Docker daemon or a remote
	// hosted builder.
	Kind BuilderKind

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

	// BuildEnv is a map of build-time environment variables to use.
	BuildEnv map[string]string
}

type DockerfileConfig struct {
	Builder string
	Root    string
	Args    Args
}

// Builder implements an image builder.
type Builder struct {
	root     string
	name     string
	args     Args
	auth     *RegistryAuth
	buildEnv map[string]string
	client   *client.Client
}

// New returns a new local builder with c.
func New(c LocalConfig) (*Builder, error) {
	if !filepath.IsAbs(c.Root) {
		return nil, fmt.Errorf("build: expected an absolute path, got %q", c.Root)
	}

	if c.Builder == "" {
		c.Builder = "manual"
	}

	if c.Args == nil {
		c.Args = make(Args)
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
		root:     c.Root,
		name:     c.Builder,
		args:     c.Args,
		auth:     c.Auth,
		buildEnv: c.BuildEnv,
		client:   client,
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
func (b *Builder) Build(ctx context.Context, taskID, version string) (*Response, error) {
	var repo = b.auth.Repo
	var name = "task-" + sanitizeTaskID(taskID)
	var uri = repo + "/" + name + ":" + version

	tree, err := NewTree()
	if err != nil {
		return nil, errors.Wrap(err, "new tree")
	}
	defer tree.Close()

	dockerfile, err := BuildDockerfile(DockerfileConfig{
		Builder: b.name,
		Root:    b.root,
		Args:    b.args,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating dockerfile")
	}
	logger.Debug(strings.TrimSpace(dockerfile))

	if err := tree.Write("Dockerfile", strings.NewReader(dockerfile)); err != nil {
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

	buildArgs := make(map[string]*string)
	for k, v := range b.buildEnv {
		value := v
		buildArgs[k] = &value
	}

	opts := types.ImageBuildOptions{
		Tags:        []string{uri},
		BuildArgs:   buildArgs,
		Platform:    "linux/amd64",
		AuthConfigs: b.authconfigs(),
	}

	resp, err := b.client.ImageBuild(ctx, bc, opts)
	if err != nil {
		return nil, errors.Wrap(err, "image build")
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var event *dockerJSONMessage.JSONMessage
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, errors.Wrap(err, "unmarshalling docker build event")
		}

		if err := event.Display(os.Stderr, isatty.IsTerminal(os.Stderr.Fd())); err != nil {
			return nil, errors.Wrap(err, "docker build")
		}
	}

	return &Response{
		ImageURL: uri,
	}, nil
}

// Push pushes the given image.
func (b *Builder) Push(ctx context.Context, uri string) error {
	authjson, err := json.Marshal(b.registryAuth())
	if err != nil {
		return err
	}

	resp, err := b.client.ImagePush(ctx, uri, types.ImagePushOptions{
		RegistryAuth: base64.URLEncoding.EncodeToString(authjson),
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		var event *dockerJSONMessage.JSONMessage
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return errors.Wrap(err, "unmarshalling docker build event")
		}

		if err := event.Display(os.Stderr, isatty.IsTerminal(os.Stderr.Fd())); err != nil {
			return errors.Wrap(err, "docker push")
		}
	}

	return nil
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

// exist ensures that all paths exists or returns an error.
func exist(paths ...string) error {
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("build: the file %s is required", path.Base(p))
		}
	}
	return nil
}

type BuilderName string

const (
	BuilderNameGo     BuilderName = "go"
	BuilderNameDeno   BuilderName = "deno"
	BuilderNamePython BuilderName = "python"
	BuilderNameNode   BuilderName = "node"
	BuilderNameDocker BuilderName = "docker"
)

func NeedsBuilding(kind api.TaskKind) bool {
	switch BuilderName(kind) {
	case BuilderNameGo, BuilderNameDeno, BuilderNamePython, BuilderNameNode, BuilderNameDocker:
		return true
	default:
		return false
	}
}

func BuildDockerfile(c DockerfileConfig) (string, error) {
	switch BuilderName(c.Builder) {
	case BuilderNameGo:
		return golang(c.Root, c.Args)
	case BuilderNameDeno:
		return deno(c.Root, c.Args)
	case BuilderNamePython:
		return python(c.Root, c.Args)
	case BuilderNameNode:
		return node(c.Root, c.Args)
	case BuilderNameDocker:
		return docker(c.Root, c.Args)
	default:
		return "", errors.Errorf("build: unknown builder type %q", c.Builder)
	}
}
