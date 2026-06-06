package config

import "testing"

func TestPlanGateDefaultsInactive(t *testing.T) {
	g := WorkPlanGateConfig{}
	applyPlanGateDefaults(&g)
	if g.IsActive() {
		t.Fatal("plan gate should default inactive")
	}
	if !g.WarnAdvisory() {
		t.Fatal("warn_is_advisory should default true")
	}
}

func TestPlanGateAgentFallback(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.applyDefaults("test")
	if cfg.PlanGateAgent() != DefaultAgentReviewer {
		t.Fatalf("plan gate agent: got %q want %q", cfg.PlanGateAgent(), DefaultAgentReviewer)
	}
	cfg.Work.Gates.Plan.Agent = "architect"
	if cfg.PlanGateAgent() != "architect" {
		t.Fatalf("explicit plan gate agent: got %q", cfg.PlanGateAgent())
	}
}

func TestPlanGateAgentLegacyFallback(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.Work.Governance.PlanGate.Agent = "legacy-architect"
	cfg.applyDefaults("test")
	if cfg.PlanGateAgent() != "legacy-architect" {
		t.Fatalf("legacy plan_gate agent: got %q", cfg.PlanGateAgent())
	}
}

func TestPlanGateWarnNonAdvisory(t *testing.T) {
	f := false
	g := WorkPlanGateConfig{WarnIsAdvisory: &f}
	if g.WarnAdvisory() {
		t.Fatal("expected non-advisory warn")
	}
}

func minimalTestConfig() *Config {
	return &Config{
		Work: WorkConfig{
			DefaultReviewer: DefaultAgentReviewer,
			Governance:      WorkGovernanceConfig{},
		},
	}
}
