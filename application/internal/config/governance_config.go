package config

import "strings"

const (
	GovernanceModeOff     = "off"
	GovernanceModePerTask = "per-task"
)

// applyWorkGovernanceDefaults fills zero-value fields on the legacy work.governance block.
func applyWorkGovernanceDefaults(g *WorkGovernanceConfig) {
	if g == nil {
		return
	}
	if strings.TrimSpace(g.Mode) == "" {
		g.Mode = GovernanceModeOff
	}
	if g.MaxRetries == nil {
		v := 2
		g.MaxRetries = &v
	}
	applyPlanGateDefaults(&g.PlanGate)
}

// GovernanceAgent returns the logical agent id for governance validation.
func (c *Config) GovernanceAgent() string {
	if c == nil {
		return DefaultAgentReviewer
	}
	if a := strings.TrimSpace(c.Work.Gates.Governance.Agent); a != "" {
		return a
	}
	return c.WorkReviewerAgent()
}
