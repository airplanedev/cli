package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var (
	// ErrMissing is returned when the config file does not exist.
	ErrMissing = errors.New("conf: config file does not exist")
)

// Config represents the configuration.
type Config struct {
	Tokens          map[string]string `json:"tokens,omitempty"`
	EnableTelemetry *bool             `json:"enableTelemetry,omitempty"`
}

// Path returns the default config path.
func path() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		// TODO(amir): friendly output.
		panic("$HOME environment variable must be set")
	}
	return filepath.Join(
		homedir,
		".airplane",
		"config",
	)
}

// ReadDefault reads the configuration from the default location.
func ReadDefault() (Config, error) {
	return Read(path())
}

// Read reads the configuration from `path`.
func Read(path string) (Config, error) {
	var cfg Config

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, ErrMissing
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, errors.Wrap(err, "read config")
	}

	if err := json.Unmarshal(buf, &cfg); err != nil {
		return cfg, errors.Wrap(err, "unmarshal config")
	}

	return cfg, nil
}

// Write writes the configuration to the given path.
func Write(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	buf, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		return errors.Wrap(err, "marshal config")
	}

	if err := ioutil.WriteFile(path, buf, 0600); err != nil {
		return errors.Wrap(err, "write config")
	}

	return nil
}

// WriteDefault saves configuration to the default location.
func WriteDefault(cfg Config) error {
	return Write(path(), cfg)
}
