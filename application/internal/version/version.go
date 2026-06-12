package version

import "fmt"

// Version, Commit, and Date are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// String returns human-readable version metadata for `asa version`.
func String() string {
	return fmt.Sprintf("Asagiri v%s\ncommit: %s\nbuilt: %s", Version, Commit, Date)
}
