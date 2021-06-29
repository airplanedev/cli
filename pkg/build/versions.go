package build

import (
	_ "embed"
	"encoding/json"

	"github.com/pkg/errors"
)

//go:embed versions.json
var versionsJSON []byte

// Versions contains a mapping table of (builder, version) to
// (node, tag, digest) image tuples. The digests are always for
// images built for the linux/amd64 architecture.
//
// This lookup table is used to construct Dockerfiles that always
// pull from the most-up-date version of the underlying base image
// based on what we have cached in the Airplane registry for customers.
type Versions map[string]map[string]Version

type Version struct {
	Image  string `json:"image"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

func (v Version) String() string {
	if v.Image == "" || v.Digest == "" {
		return ""
	}

	return v.Image + "@" + v.Digest
}

func GetVersions() (Versions, error) {
	var versions Versions
	if err := json.Unmarshal(versionsJSON, &versions); err != nil {
		return Versions{}, errors.Wrap(err, "unmarshalling versions.json")
	}

	return versions, nil
}

func GetVersion(builder Name, version string) (Version, error) {
	versions, err := GetVersions()
	if err != nil {
		return Version{}, err
	}

	builderVersions, ok := versions[string(builder)]
	if !ok {
		return Version{}, errors.Errorf("unknown builder: %s", builder)
	}

	return builderVersions[version], nil
}
