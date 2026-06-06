package config_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestDefaultAgentConstantsAreLogical(t *testing.T) {
	require.Equal(t, "laprogrammerie", config.DefaultAgentSpec)
	require.Equal(t, "dev", config.DefaultAgentDev)
	require.Equal(t, "reviewer", config.DefaultAgentReviewer)
	require.Equal(t, "local-rag", config.DefaultAgentEnrich)
}

func TestApplyDefaultsUsesLogicalConstants(t *testing.T) {
	cfg := config.NewTestConfig("myproject")
	require.Equal(t, config.DefaultAgentSpec, cfg.Work.DefaultSpecAgent)
	require.Equal(t, config.DefaultAgentDev, cfg.Work.DefaultAgent)
	require.Equal(t, config.DefaultAgentReviewer, cfg.Work.DefaultReviewer)
	require.Equal(t, config.DefaultAgentEnrich, cfg.Work.DefaultEnricher)
}

func TestNewTestConfigBootstrapCatalog(t *testing.T) {
	cfg := config.NewTestConfig("demo")

	for _, id := range config.DefaultLogicalAgentIDs() {
		_, err := cfg.LookupAgent(id)
		require.NoError(t, err, "agent %q", id)
		typ, merged, err := cfg.MergedAgentRuntime(id)
		require.NoError(t, err, "runtime %q", id)
		require.NotEmpty(t, typ)
		require.NotEmpty(t, merged.Command)
	}

	require.Contains(t, cfg.Providers, "kiro-cli")
	require.Contains(t, cfg.Providers, "claude-code")
	require.Contains(t, cfg.Providers, "ollama")
	require.Equal(t, "kiro-cli", cfg.Agents["dev"].Provider)
	require.Equal(t, "claude-code", cfg.Agents["reviewer"].Provider)
	require.Equal(t, "ollama", cfg.Agents["local-rag"].Provider)
}

func TestWorkAgentHelpersPreferConfigWorkSection(t *testing.T) {
	cfg := &config.Config{
		Work: config.WorkConfig{
			DefaultSpecAgent: "kiro",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
			DefaultEnricher:  "ollama",
		},
	}
	require.Equal(t, "kiro", cfg.WorkSpecAgent())
	require.Equal(t, "cursor", cfg.WorkDevAgent())
	require.Equal(t, "codex", cfg.WorkReviewerAgent())
	require.Equal(t, "ollama", cfg.WorkEnricherAgent())
}

func TestApplyRecommendedRuntimeCatalogSkipsExistingAgents(t *testing.T) {
	cfg := &config.Config{}
	cfg.Agents = map[string]config.Agent{
		"cursor": {Command: "cursor-agent"},
	}
	cfg.Work.DefaultAgent = "cursor"
	config.ApplyRecommendedRuntimeCatalog(cfg)
	require.Len(t, cfg.Providers, 0)
	require.Equal(t, "cursor-agent", cfg.Agents["cursor"].Command)
}
