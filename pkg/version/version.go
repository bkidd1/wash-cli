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

// Info holds version information
type Info struct {
	Version   string
	BuildDate string
	GitCommit string
	GoVersion string
	Platform  string
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		BuildDate: BuildDate,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("wash version %s\nBuild Date: %s\nGit Commit: %s\nGo Version: %s\nPlatform: %s",
		i.Version, i.BuildDate, i.GitCommit, i.GoVersion, i.Platform)
}
