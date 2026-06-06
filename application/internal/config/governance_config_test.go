package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkGovernanceDefaults(t *testing.T) {
	g := WorkGovernanceConfig{}
	applyWorkGovernanceDefaults(&g)
	if g.Enabled {
		t.Fatal("enabled should default false")
	}
	if g.Mode != GovernanceModeOff {
		t.Fatalf("mode: got %q want off", g.Mode)
	}
	if g.MaxRetries == nil || *g.MaxRetries != 2 {
		got := 0
		if g.MaxRetries != nil {
			got = *g.MaxRetries
		}
		t.Fatalf("max_retries: got %d want 2", got)
	}
	if !g.WarnAdvisory() {
		t.Fatal("warn_is_advisory should default true")
	}
}

func TestWorkGovernanceMaxRetriesZeroExplicit(t *testing.T) {
	zero := 0
	g := WorkGovernanceConfig{MaxRetries: &zero}
	applyWorkGovernanceDefaults(&g)
	if g.MaxRetries == nil || *g.MaxRetries != 0 {
		got := -1
		if g.MaxRetries != nil {
			got = *g.MaxRetries
		}
		t.Fatalf("explicit max_retries=0: got %d want 0", got)
	}
	if g.MaxRetriesValue() != 0 {
		t.Fatalf("MaxRetriesValue: got %d want 0", g.MaxRetriesValue())
	}
}

func TestWorkGovernanceIsActive(t *testing.T) {
	off := WorkGovernanceConfig{Enabled: true, Mode: GovernanceModeOff}
	if off.IsActive() {
		t.Fatal("mode off should not be active")
	}
	perTask := WorkGovernanceConfig{Enabled: true, Mode: GovernanceModePerTask}
	if !perTask.IsActive() {
		t.Fatal("enabled per-task should be active")
	}
	disabled := WorkGovernanceConfig{Enabled: false, Mode: GovernanceModePerTask}
	if disabled.IsActive() {
		t.Fatal("disabled should not be active")
	}
}

func TestGovernanceAgentDefaultReviewer(t *testing.T) {
	cfg := &Config{}
	if cfg.GovernanceAgent() != DefaultAgentReviewer {
		t.Fatalf("governance agent: got %q want %q", cfg.GovernanceAgent(), DefaultAgentReviewer)
	}
	cfg.Work.Governance.Agent = "architect"
	if cfg.GovernanceAgent() != "architect" {
		t.Fatalf("explicit agent: got %q", cfg.GovernanceAgent())
	}
}

func TestWarnAdvisoryExplicitFalse(t *testing.T) {
	f := false
	g := WorkGovernanceConfig{WarnIsAdvisory: &f}
	if g.WarnAdvisory() {
		t.Fatal("warn_is_advisory false should not be advisory")
	}
}

func TestGovernanceEnabledButInactive(t *testing.T) {
	g := WorkGovernanceConfig{Enabled: true, Mode: "smart"}
	if g.IsActive() {
		t.Fatal("smart mode should not be active")
	}
	if !g.EnabledButInactive() {
		t.Fatal("enabled smart should report inactive")
	}
	off := WorkGovernanceConfig{Enabled: true, Mode: GovernanceModeOff}
	if off.EnabledButInactive() {
		t.Fatal("enabled off should not warn inactive")
	}
}

func TestLegacyConfigWithoutGovernanceBlockIsInactive(t *testing.T) {
	const raw = `
project:
  name: legacy
work:
  default_agent: dev
  default_reviewer: reviewer
`
	var c Config
	if err := yaml.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	c.applyDefaults("legacy")
	c.applyV3Defaults()

	if c.Work.Governance.Enabled {
		t.Fatal("enabled should default false when governance block omitted")
	}
	if c.Work.Governance.IsActive() {
		t.Fatal("governance must stay inactive without explicit enable")
	}
	if c.Work.Governance.MaxRetriesValue() != 2 {
		t.Fatalf("max_retries default: got %d want 2", c.Work.Governance.MaxRetriesValue())
	}
	if c.GovernanceAgent() != DefaultAgentReviewer {
		t.Fatalf("governance agent fallback: got %q", c.GovernanceAgent())
	}
}
