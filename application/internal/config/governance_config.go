package config

import "strings"

const (
	GovernanceModeOff     = "off"
	GovernanceModePerTask = "per-task"
)

// applyWorkGovernanceDefaults fills zero-value governance fields (enabled stays false).
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
}

// MaxRetriesValue returns configured max_retries (default 2 when unset).
// Zero is valid: no relance after the first governance FAIL.
func (g WorkGovernanceConfig) MaxRetriesValue() int {
	if g.MaxRetries == nil {
		return 2
	}
	return *g.MaxRetries
}

// IsActive reports whether per-task governance gates run after dev.
func (g WorkGovernanceConfig) IsActive() bool {
	if !g.Enabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.Mode), GovernanceModePerTask)
}

// WarnAdvisory reports whether WARN verdicts allow the workflow to continue (default true).
func (g WorkGovernanceConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}

// EnabledButInactive reports enabled=true with a mode other than per-task (gate skipped).
func (g WorkGovernanceConfig) EnabledButInactive() bool {
	if !g.Enabled || g.IsActive() {
		return false
	}
	mode := strings.TrimSpace(g.Mode)
	return mode != "" && !strings.EqualFold(mode, GovernanceModeOff)
}

// GovernanceAgent returns the logical agent id for governance validation.
func (c *Config) GovernanceAgent() string {
	if c == nil {
		return DefaultAgentReviewer
	}
	if a := strings.TrimSpace(c.Work.Governance.Agent); a != "" {
		return a
	}
	return c.WorkReviewerAgent()
}
