package config

import "strings"

func applyPlanGateDefaults(g *WorkPlanGateConfig) {
	if g == nil {
		return
	}
}

// IsActive reports whether the plan gate runs after PlanFeature.
func (g WorkPlanGateConfig) IsActive() bool {
	return g.Enabled
}

// WarnAdvisory reports whether WARN verdicts allow planning to continue (default true).
func (g WorkPlanGateConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}

// PlanGateAgent returns the logical agent id for plan validation.
func (c *Config) PlanGateAgent() string {
	if c == nil {
		return DefaultAgentReviewer
	}
	if a := strings.TrimSpace(c.Work.Gates.Plan.Agent); a != "" {
		return a
	}
	return c.GovernanceAgent()
}
