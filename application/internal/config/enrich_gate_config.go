package config

import "strings"

// Default enrich gate finding codes that map to FAIL when classified (MVP).
var defaultEnrichGateFailOn = []string{
	"missing_files_scope",
	"invalid_validation_commands",
	"enrichment_not_actionable",
	"enrich_gate_parse_error",
}

// WorkEnrichGateConfig controls read-only validation after enrich, before dev (enrich-gate).
type WorkEnrichGateConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Mode           string   `yaml:"mode"` // off | per-task
	Agent          string   `yaml:"agent"`
	FailOn         []string `yaml:"fail_on"`
	WarnIsAdvisory *bool    `yaml:"warn_is_advisory"`
}

func applyEnrichGateDefaults(g *WorkEnrichGateConfig) {
	if g == nil {
		return
	}
	if strings.TrimSpace(g.Mode) == "" {
		g.Mode = GovernanceModeOff
	}
	if g.FailOn == nil {
		g.FailOn = append([]string(nil), defaultEnrichGateFailOn...)
	}
}

// DefaultEnrichGateFailOn returns a copy of the default fail_on codes for enrich gate.
func DefaultEnrichGateFailOn() []string {
	return append([]string(nil), defaultEnrichGateFailOn...)
}

func (g WorkEnrichGateConfig) IsActive() bool {
	if !g.Enabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.Mode), GovernanceModePerTask)
}

func (g WorkEnrichGateConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}

func (g WorkEnrichGateConfig) EnabledButInactive() bool {
	if !g.Enabled || g.IsActive() {
		return false
	}
	mode := strings.TrimSpace(g.Mode)
	return mode != "" && !strings.EqualFold(mode, GovernanceModeOff)
}

// EnrichGateAgent returns the logical agent id for enrich gate validation.
func (c *Config) EnrichGateAgent() string {
	if c == nil {
		return DefaultAgentReviewer
	}
	if a := strings.TrimSpace(c.Work.Gates.Enrich.Agent); a != "" {
		return a
	}
	return c.GovernanceAgent()
}
