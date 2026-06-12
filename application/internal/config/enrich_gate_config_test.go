package config

import (
	"slices"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkEnrichGateDefaults(t *testing.T) {
	g := WorkEnrichGateConfig{}
	applyEnrichGateDefaults(&g)
	if g.Mode != GovernanceModeOff {
		t.Fatalf("mode default: got %q", g.Mode)
	}
	if g.IsActive() {
		t.Fatal("enrich gate should be inactive when enabled=false")
	}
	if !g.WarnAdvisory() {
		t.Fatal("warn should be advisory by default")
	}
	wantFailOn := DefaultEnrichGateFailOn()
	if !slices.Equal(g.FailOn, wantFailOn) {
		t.Fatalf("fail_on default: got %v want %v", g.FailOn, wantFailOn)
	}
	if slices.Contains(g.FailOn, "empty_context_when_required") {
		t.Fatal("empty_context_when_required must not be in default fail_on")
	}
}

func TestWorkEnrichGateIsActive(t *testing.T) {
	active := WorkEnrichGateConfig{Enabled: true, Mode: GovernanceModePerTask}
	applyEnrichGateDefaults(&active)
	if !active.IsActive() {
		t.Fatal("expected active with enabled + per-task")
	}
	off := WorkEnrichGateConfig{Enabled: true, Mode: GovernanceModeOff}
	applyEnrichGateDefaults(&off)
	if off.IsActive() {
		t.Fatal("mode off should be inactive")
	}
	disabled := WorkEnrichGateConfig{Enabled: false, Mode: GovernanceModePerTask}
	if disabled.IsActive() {
		t.Fatal("enabled=false should be inactive")
	}
}

func TestWorkEnrichGateEnabledButInactive(t *testing.T) {
	g := WorkEnrichGateConfig{Enabled: true, Mode: "smart"}
	applyEnrichGateDefaults(&g)
	if g.IsActive() {
		t.Fatal("unexpected active for unsupported mode")
	}
	if !g.EnabledButInactive() {
		t.Fatal("expected enabledButInactive for unsupported mode")
	}
}

func TestWorkEnrichGateFailOnPreservesExplicitEmpty(t *testing.T) {
	g := WorkEnrichGateConfig{FailOn: []string{}}
	applyEnrichGateDefaults(&g)
	if len(g.FailOn) != 0 {
		t.Fatalf("explicit empty fail_on should stay empty, got %v", g.FailOn)
	}
}

func TestWorkEnrichGateFailOnCustomOverridesDefault(t *testing.T) {
	g := WorkEnrichGateConfig{FailOn: []string{"custom_code"}}
	applyEnrichGateDefaults(&g)
	if !slices.Equal(g.FailOn, []string{"custom_code"}) {
		t.Fatalf("custom fail_on: got %v", g.FailOn)
	}
}

func TestEnrichGateAgentFallback(t *testing.T) {
	cfg := minimalTestConfig()
	cfg.applyDefaults("test")
	if cfg.EnrichGateAgent() != DefaultAgentReviewer {
		t.Fatalf("enrich gate agent: got %q want %q", cfg.EnrichGateAgent(), DefaultAgentReviewer)
	}
	cfg.Work.Gates.Enrich.Agent = "architect"
	if cfg.EnrichGateAgent() != "architect" {
		t.Fatalf("explicit enrich gate agent: got %q", cfg.EnrichGateAgent())
	}
}

func TestWorkGatesEnrichInYAML(t *testing.T) {
	var c Config
	if err := yaml.Unmarshal([]byte(`
work:
  gates:
    enrich:
      enabled: true
      mode: per-task
`), &c); err != nil {
		t.Fatal(err)
	}
	normalizeWorkGates(&c.Work)
	if !c.Work.Gates.Enrich.IsActive() {
		t.Fatal("expected enrich gate active from yaml")
	}
	if !slices.Contains(c.Work.Gates.Enrich.FailOn, "missing_files_scope") {
		t.Fatal("expected default fail_on after normalize")
	}
}

func TestNormalizeWorkGatesEnrichDefaultsWithoutYAML(t *testing.T) {
	var c Config
	c.applyDefaults("test")
	if c.Work.Gates.Enrich.IsActive() {
		t.Fatal("expected enrich inactive after applyDefaults")
	}
	if c.Work.Gates.Enrich.Mode != GovernanceModeOff {
		t.Fatalf("mode: got %q", c.Work.Gates.Enrich.Mode)
	}
}

func TestWorkEnrichGateWarnNonAdvisory(t *testing.T) {
	f := false
	g := WorkEnrichGateConfig{WarnIsAdvisory: &f}
	if g.WarnAdvisory() {
		t.Fatal("expected non-advisory warn")
	}
}
