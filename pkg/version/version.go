// Package version includes the version information.
package version

var (
	// Raw is the string representation of the version. This will be replaced
	// with the calculated version at build time.
	// Set via LDFLAGS in hack/build.sh.
	Raw = "was not built with version info"

	// Commit is the commit hash from which the agent-installer-utils was built.
	// Set via LDFLAGS in hack/build.sh.
	Commit = ""
)
