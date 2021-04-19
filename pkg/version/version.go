package version

// Set by Go Releaser.
var (
	version string = "<unknown>"
	date    string = "<unknown>"
)

func Get() string {
	return version
}

func Date() string {
	return date
}
