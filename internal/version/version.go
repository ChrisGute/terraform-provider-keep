// Package version contains version information for the provider
package version

import "fmt"

var (
	// Version is the current version of the provider
	Version = "0.1.0"

	// GitCommit is the git commit hash of the release
	GitCommit = ""
)

// String returns a human-readable version string
func String() string {
	if GitCommit != "" {
		return fmt.Sprintf("%s+%s", Version, GitCommit)
	}
	return Version
}
