package config

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkGatesConfig groups validation gates under work (plan, governance, enrich, human_review, verify_evidence, …).
type WorkGatesConfig struct {
	Plan           WorkPlanGateConfig           `yaml:"plan"`
	Governance     WorkGovernanceGateConfig     `yaml:"governance"`
	Enrich         WorkEnrichGateConfig         `yaml:"enrich"`
	HumanReview    WorkHumanReviewGateConfig    `yaml:"human_review"`
	VerifyEvidence WorkVerifyEvidenceGateConfig `yaml:"verify_evidence"`
	Trust          WorkTrustGateConfig          `yaml:"trust"`

	planPresent       bool
	governancePresent bool
}

// WorkGovernanceGateConfig controls post-dev governance gates (task-validation-gates).
type WorkGovernanceGateConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Mode           string   `yaml:"mode"` // off | per-task
	Agent          string   `yaml:"agent"`
	FailOn         []string `yaml:"fail_on"`
	WarnIsAdvisory *bool    `yaml:"warn_is_advisory"`
	MaxRetries     *int     `yaml:"max_retries"`
}

// UnmarshalYAML records explicit presence of plan/governance keys so legacy merge is skipped.
func (g *WorkGatesConfig) UnmarshalYAML(value *yaml.Node) error {
	g.planPresent = false
	g.governancePresent = false
	if value != nil {
		for i := 0; i+1 < len(value.Content); i += 2 {
			switch value.Content[i].Value {
			case "plan":
				g.planPresent = true
			case "governance":
				g.governancePresent = true
			}
		}
	}
	type plain struct {
		Plan           WorkPlanGateConfig           `yaml:"plan"`
		Governance     WorkGovernanceGateConfig     `yaml:"governance"`
		Enrich         WorkEnrichGateConfig         `yaml:"enrich"`
		HumanReview    WorkHumanReviewGateConfig    `yaml:"human_review"`
		VerifyEvidence WorkVerifyEvidenceGateConfig `yaml:"verify_evidence"`
		Trust          WorkTrustGateConfig          `yaml:"trust"`
	}
	var p plain
	if err := value.Decode(&p); err != nil {
		return err
	}
	g.Plan = p.Plan
	g.Governance = p.Governance
	g.Enrich = p.Enrich
	g.HumanReview = p.HumanReview
	g.VerifyEvidence = p.VerifyEvidence
	g.Trust = p.Trust
	return nil
}

func applyWorkGatesDefaults(g *WorkGatesConfig) {
	if g == nil {
		return
	}
	applyWorkGovernanceGateDefaults(&g.Governance)
	applyPlanGateDefaults(&g.Plan)
	applyEnrichGateDefaults(&g.Enrich)
	applyHumanReviewGateDefaults(&g.HumanReview)
	applyVerifyEvidenceGateDefaults(&g.VerifyEvidence)
	applyTrustGateDefaults(&g.Trust)
}

func applyWorkGovernanceGateDefaults(g *WorkGovernanceGateConfig) {
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

func (g WorkGovernanceGateConfig) MaxRetriesValue() int {
	if g.MaxRetries == nil {
		return 2
	}
	return *g.MaxRetries
}

func (g WorkGovernanceGateConfig) IsActive() bool {
	if !g.Enabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.Mode), GovernanceModePerTask)
}

func (g WorkGovernanceGateConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}

func (g WorkGovernanceGateConfig) EnabledButInactive() bool {
	if !g.Enabled || g.IsActive() {
		return false
	}
	mode := strings.TrimSpace(g.Mode)
	return mode != "" && !strings.EqualFold(mode, GovernanceModeOff)
}

// NormalizeWorkGates merges legacy work.governance into work.gates and applies gate defaults.
func NormalizeWorkGates(w *WorkConfig) {
	normalizeWorkGates(w)
}

func normalizeWorkGates(w *WorkConfig) {
	if w == nil {
		return
	}
	if !w.Gates.governancePresent && legacyGovernanceCoreConfigured(w.Governance) {
		w.Gates.Governance = WorkGovernanceGateConfig{
			Enabled:        w.Governance.Enabled,
			Mode:           w.Governance.Mode,
			Agent:          w.Governance.Agent,
			FailOn:         w.Governance.FailOn,
			WarnIsAdvisory: w.Governance.WarnIsAdvisory,
			MaxRetries:     w.Governance.MaxRetries,
		}
	}
	if !w.Gates.planPresent && planGateConfigured(w.Governance.PlanGate) {
		w.Gates.Plan = w.Governance.PlanGate
	}
	applyWorkGatesDefaults(&w.Gates)
}

func planGateConfigured(g WorkPlanGateConfig) bool {
	return g.Enabled ||
		strings.TrimSpace(g.Agent) != "" ||
		len(g.FailOn) > 0 ||
		g.WarnIsAdvisory != nil
}

func legacyGovernanceCoreConfigured(g WorkGovernanceConfig) bool {
	return g.Enabled ||
		strings.TrimSpace(g.Mode) != "" ||
		strings.TrimSpace(g.Agent) != "" ||
		len(g.FailOn) > 0 ||
		g.WarnIsAdvisory != nil ||
		g.MaxRetries != nil
}
