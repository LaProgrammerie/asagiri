package config

import "strings"

var defaultTrustGateBlockVerdicts = []string{"blocked"}
var defaultTrustGateWarnVerdicts = []string{"risky"}

// WorkTrustGateConfig controls read-only trust synthesis enforcement after verify evidence, before review.
type WorkTrustGateConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Mode           string   `yaml:"mode"` // off | per-task
	MinScore       *float64 `yaml:"min_score"`
	BlockVerdicts  []string `yaml:"block_verdicts"`
	WarnVerdicts   []string `yaml:"warn_verdicts"`
	WarnIsAdvisory *bool    `yaml:"warn_is_advisory"`
}

func applyTrustGateDefaults(g *WorkTrustGateConfig) {
	if g == nil {
		return
	}
	if strings.TrimSpace(g.Mode) == "" {
		g.Mode = GovernanceModeOff
	}
	if g.MinScore == nil {
		v := 70.0
		g.MinScore = &v
	}
	if g.BlockVerdicts == nil {
		g.BlockVerdicts = append([]string(nil), defaultTrustGateBlockVerdicts...)
	}
	if g.WarnVerdicts == nil {
		g.WarnVerdicts = append([]string(nil), defaultTrustGateWarnVerdicts...)
	}
}

// DefaultTrustGateBlockVerdicts returns a copy of default block_verdicts.
func DefaultTrustGateBlockVerdicts() []string {
	return append([]string(nil), defaultTrustGateBlockVerdicts...)
}

// DefaultTrustGateWarnVerdicts returns a copy of default warn_verdicts.
func DefaultTrustGateWarnVerdicts() []string {
	return append([]string(nil), defaultTrustGateWarnVerdicts...)
}

func (g WorkTrustGateConfig) IsActive() bool {
	if !g.Enabled {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(g.Mode), GovernanceModePerTask)
}

func (g WorkTrustGateConfig) WarnAdvisory() bool {
	if g.WarnIsAdvisory == nil {
		return true
	}
	return *g.WarnIsAdvisory
}

func (g WorkTrustGateConfig) EnabledButInactive() bool {
	if !g.Enabled || g.IsActive() {
		return false
	}
	mode := strings.TrimSpace(g.Mode)
	return mode != "" && !strings.EqualFold(mode, GovernanceModeOff)
}

func (g WorkTrustGateConfig) MinScoreValue() float64 {
	if g.MinScore == nil {
		return 70
	}
	return *g.MinScore
}
