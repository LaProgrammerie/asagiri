package config

import "strings"

// WorkHumanReviewGateConfig controls manual human validation after dev (and governance when active).
type WorkHumanReviewGateConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Mode           string   `yaml:"mode"` // off | per-task
	VerdictFile    string   `yaml:"verdict_file"`
	FailOn         []string `yaml:"fail_on"`
	WarnIsAdvisory *bool    `yaml:"warn_is_advisory"`
}

func applyHumanReviewGateDefaults(g *WorkHumanReviewGateConfig) {
	if g == nil {
		return
	}
	if strings.TrimSpace(g.Mode) == "" {
		g.Mode = GovernanceModeOff
	}
}

func (g WorkHumanReviewGateConfig) IsActive() bool {
	if !g.Enabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.Mode), GovernanceModePerTask)
}

func (g WorkHumanReviewGateConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}
