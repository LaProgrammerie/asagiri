package agentcontext

import (
	"path/filepath"
	"strings"
)

// AgentLogDirForPhase returns the agent log directory for a workflow phase.
// Phase dev keeps the historical layout (.asagiri/logs/<task>/agents/<agent>/).
func AgentLogDirForPhase(repoRoot, taskID, agentID, phase string) string {
	base := AgentLogDir(repoRoot, taskID, agentID)
	p := strings.TrimSpace(phase)
	if p == "" || p == "dev" {
		return base
	}
	return filepath.Join(base, p)
}
