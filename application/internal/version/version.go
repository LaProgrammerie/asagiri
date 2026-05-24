package version

import "fmt"

// Version, Commit, and Date are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// String returns human-readable version metadata for `agentflow version`.
func String() string {
	return fmt.Sprintf("AgentFlow v%s\ncommit: %s\nbuilt: %s", Version, Commit, Date)
}
