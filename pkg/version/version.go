package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of the CLI
	Version = "dev"
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
	// GitCommit is the git commit hash
	GitCommit = "unknown"
	// GoVersion is the version of Go used to build the binary
	GoVersion = runtime.Version()
)

// String returns a formatted string with version information
func String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)", Version, GitCommit, BuildDate, GoVersion)
}
