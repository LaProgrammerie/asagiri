package config

import "strings"

const (
	AgentObservabilityModeBestEffort = "best_effort"
	AgentObservabilityModeWarn       = "warn"
	AgentObservabilityModeStrict     = "strict"
)

// WorkAgentObservabilityConfig controls how agent ledger/logs/contract write failures are handled.
type WorkAgentObservabilityConfig struct {
	Mode string `yaml:"mode"` // best_effort | warn | strict
}

// NormalizeWorkAgentObservability applies defaults for work.agent_observability.
func NormalizeWorkAgentObservability(o *WorkAgentObservabilityConfig) {
	normalizeAgentObservability(o)
}

func normalizeAgentObservability(o *WorkAgentObservabilityConfig) {
	if o == nil {
		return
	}
	mode := strings.TrimSpace(o.Mode)
	if mode == "" {
		o.Mode = AgentObservabilityModeBestEffort
		return
	}
	switch mode {
	case AgentObservabilityModeBestEffort, AgentObservabilityModeWarn, AgentObservabilityModeStrict:
		o.Mode = mode
	default:
		o.Mode = AgentObservabilityModeBestEffort
	}
}
