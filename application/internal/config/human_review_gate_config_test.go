package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkHumanReviewDefaults(t *testing.T) {
	g := WorkHumanReviewGateConfig{}
	applyHumanReviewGateDefaults(&g)
	if g.Mode != GovernanceModeOff {
		t.Fatalf("mode default: got %q", g.Mode)
	}
	if g.IsActive() {
		t.Fatal("human_review should be inactive when enabled=false")
	}
	if !g.WarnAdvisory() {
		t.Fatal("warn should be advisory by default")
	}
}

func TestWorkHumanReviewIsActive(t *testing.T) {
	active := WorkHumanReviewGateConfig{Enabled: true, Mode: GovernanceModePerTask}
	if !active.IsActive() {
		t.Fatal("expected active")
	}
	off := WorkHumanReviewGateConfig{Enabled: true, Mode: GovernanceModeOff}
	if off.IsActive() {
		t.Fatal("mode off should be inactive")
	}
}

func TestWorkGatesHumanReviewInYAML(t *testing.T) {
	var c Config
	if err := yaml.Unmarshal([]byte(`
work:
  gates:
    human_review:
      enabled: true
      mode: per-task
`), &c); err != nil {
		t.Fatal(err)
	}
	normalizeWorkGates(&c.Work)
	if !c.Work.Gates.HumanReview.IsActive() {
		t.Fatal("expected human_review active from yaml")
	}
}
