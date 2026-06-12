package config_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAgentObservabilityDefaultBestEffort(t *testing.T) {
	cfg := config.NewTestConfig("repo")
	require.Equal(t, config.AgentObservabilityModeBestEffort, cfg.Work.AgentObservability.Mode)
}

func TestNormalizeWorkAgentObservabilityInvalidMode(t *testing.T) {
	o := config.WorkAgentObservabilityConfig{Mode: "nope"}
	config.NormalizeWorkAgentObservability(&o)
	require.Equal(t, config.AgentObservabilityModeBestEffort, o.Mode)
}
