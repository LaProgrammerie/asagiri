package config

import (
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkVerifyEvidenceGateDefaults(t *testing.T) {
	g := WorkVerifyEvidenceGateConfig{}
	applyVerifyEvidenceGateDefaults(&g)
	if g.Mode != GovernanceModeOff {
		t.Fatalf("mode default: got %q", g.Mode)
	}
	if g.IsActive() {
		t.Fatal("verify evidence gate should be inactive when enabled=false")
	}
	if !g.WarnAdvisory() {
		t.Fatal("warn should be advisory by default")
	}
	wantFailOn := DefaultVerifyEvidenceGateFailOn()
	if !slices.Equal(g.FailOn, wantFailOn) {
		t.Fatalf("fail_on default: got %v want %v", g.FailOn, wantFailOn)
	}
}

func TestWorkVerifyEvidenceGateIsActive(t *testing.T) {
	active := WorkVerifyEvidenceGateConfig{Enabled: true, Mode: GovernanceModePerTask}
	applyVerifyEvidenceGateDefaults(&active)
	if !active.IsActive() {
		t.Fatal("expected active with enabled + per-task")
	}
	off := WorkVerifyEvidenceGateConfig{Enabled: true, Mode: GovernanceModeOff}
	applyVerifyEvidenceGateDefaults(&off)
	if off.IsActive() {
		t.Fatal("mode off should be inactive")
	}
	disabled := WorkVerifyEvidenceGateConfig{Enabled: false, Mode: GovernanceModePerTask}
	if disabled.IsActive() {
		t.Fatal("enabled=false should be inactive")
	}
}

func TestWorkVerifyEvidenceGateEnabledButInactive(t *testing.T) {
	g := WorkVerifyEvidenceGateConfig{Enabled: true, Mode: "smart"}
	applyVerifyEvidenceGateDefaults(&g)
	if g.IsActive() {
		t.Fatal("unexpected active for unsupported mode")
	}
	if !g.EnabledButInactive() {
		t.Fatal("expected enabledButInactive for unsupported mode")
	}
}

func TestWorkVerifyEvidenceGateFailOnPreservesExplicitEmpty(t *testing.T) {
	g := WorkVerifyEvidenceGateConfig{FailOn: []string{}}
	applyVerifyEvidenceGateDefaults(&g)
	if len(g.FailOn) != 0 {
		t.Fatalf("explicit empty fail_on should stay empty, got %v", g.FailOn)
	}
}

func TestWorkVerifyEvidenceGateFailOnCustomOverridesDefault(t *testing.T) {
	g := WorkVerifyEvidenceGateConfig{FailOn: []string{"custom_code"}}
	applyVerifyEvidenceGateDefaults(&g)
	if !slices.Equal(g.FailOn, []string{"custom_code"}) {
		t.Fatalf("custom fail_on: got %v", g.FailOn)
	}
}

func TestVerifyEvidenceGateAgentFallback(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.applyDefaults("test")
	if cfg.VerifyEvidenceGateAgent() != DefaultAgentReviewer {
		t.Fatalf("verify evidence gate agent: got %q want %q", cfg.VerifyEvidenceGateAgent(), DefaultAgentReviewer)
	}
	cfg.Work.Gates.VerifyEvidence.Agent = "architect"
	if cfg.VerifyEvidenceGateAgent() != "architect" {
		t.Fatalf("explicit verify evidence gate agent: got %q", cfg.VerifyEvidenceGateAgent())
	}
}

func TestWorkGatesVerifyEvidenceInYAML(t *testing.T) {
	var c Config
	if err := yaml.Unmarshal([]byte(`
work:
  gates:
    verify_evidence:
      enabled: true
      mode: per-task
`), &c); err != nil {
		t.Fatal(err)
	}
	normalizeWorkGates(&c.Work)
	if !c.Work.Gates.VerifyEvidence.IsActive() {
		t.Fatal("expected verify evidence gate active from yaml")
	}
	if !slices.Contains(c.Work.Gates.VerifyEvidence.FailOn, "missing_validation_commands") {
		t.Fatal("expected default fail_on after normalize")
	}
}

func TestNormalizeWorkGatesVerifyEvidenceDefaultsWithoutYAML(t *testing.T) {
	var c Config
	c.applyDefaults("test")
	if c.Work.Gates.VerifyEvidence.IsActive() {
		t.Fatal("expected verify evidence inactive after applyDefaults")
	}
	if c.Work.Gates.VerifyEvidence.Mode != GovernanceModeOff {
		t.Fatalf("mode: got %q", c.Work.Gates.VerifyEvidence.Mode)
	}
}

func TestWorkVerifyEvidenceGateWarnNonAdvisory(t *testing.T) {
	f := false
	g := WorkVerifyEvidenceGateConfig{WarnIsAdvisory: &f}
	if g.WarnAdvisory() {
		t.Fatal("expected non-advisory warn")
	}
}
