package config

import "strings"

// Default verify evidence gate finding codes that map to FAIL when classified (MVP).
var defaultVerifyEvidenceGateFailOn = []string{
	"missing_validation_commands",
	"validation_output_missing",
	"verification_not_actionable",
	"verify_evidence_gate_parse_error",
}

// WorkVerifyEvidenceGateConfig controls read-only validation after local verify, before verified (verify-evidence-gate).
type WorkVerifyEvidenceGateConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Mode           string   `yaml:"mode"` // off | per-task
	Agent          string   `yaml:"agent"`
	FailOn         []string `yaml:"fail_on"`
	WarnIsAdvisory *bool    `yaml:"warn_is_advisory"`
}

func applyVerifyEvidenceGateDefaults(g *WorkVerifyEvidenceGateConfig) {
	if g == nil {
		return
	}
	if strings.TrimSpace(g.Mode) == "" {
		g.Mode = GovernanceModeOff
	}
	if g.FailOn == nil {
		g.FailOn = append([]string(nil), defaultVerifyEvidenceGateFailOn...)
	}
}

// DefaultVerifyEvidenceGateFailOn returns a copy of the default fail_on codes for verify evidence gate.
func DefaultVerifyEvidenceGateFailOn() []string {
	return append([]string(nil), defaultVerifyEvidenceGateFailOn...)
}

func (g WorkVerifyEvidenceGateConfig) IsActive() bool {
	if !g.Enabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.Mode), GovernanceModePerTask)
}

func (g WorkVerifyEvidenceGateConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}

func (g WorkVerifyEvidenceGateConfig) EnabledButInactive() bool {
	if !g.Enabled || g.IsActive() {
		return false
	}
	mode := strings.TrimSpace(g.Mode)
	return mode != "" && !strings.EqualFold(mode, GovernanceModeOff)
}

// VerifyEvidenceGateAgent returns the logical agent id for verify evidence gate validation.
func (c *Config) VerifyEvidenceGateAgent() string {
	if c == nil {
		return DefaultAgentReviewer
	}
	if a := strings.TrimSpace(c.Work.Gates.VerifyEvidence.Agent); a != "" {
		return a
	}
	return c.GovernanceAgent()
}
