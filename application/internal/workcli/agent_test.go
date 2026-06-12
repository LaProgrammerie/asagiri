package workcli

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestResolveWorkAgentUsesConfigWhenFlagUnchanged(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("agent", config.DefaultAgentEnrich, "")
	cfg := &config.Config{
		Work: config.WorkConfig{DefaultEnricher: "ollama"},
	}
	got := ResolveWorkAgent(cmd, config.DefaultAgentEnrich, (*config.Config).WorkEnricherAgent, cfg)
	require.Equal(t, "ollama", got)
}

func TestResolveWorkAgentKeepsExplicitFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("agent", config.DefaultAgentEnrich, "")
	require.NoError(t, cmd.Flags().Set("agent", "custom-agent"))
	cfg := &config.Config{
		Work: config.WorkConfig{DefaultEnricher: "ollama"},
	}
	got := ResolveWorkAgent(cmd, "custom-agent", (*config.Config).WorkEnricherAgent, cfg)
	require.Equal(t, "custom-agent", got)
}
