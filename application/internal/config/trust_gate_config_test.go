package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTrustGateDefaults(t *testing.T) {
	g := WorkTrustGateConfig{}
	applyTrustGateDefaults(&g)
	require.False(t, g.Enabled)
	require.Equal(t, GovernanceModeOff, g.Mode)
	require.Equal(t, 70.0, g.MinScoreValue())
	require.Equal(t, []string{"blocked"}, g.BlockVerdicts)
	require.Equal(t, []string{"risky"}, g.WarnVerdicts)
	require.True(t, g.WarnAdvisory())
}

func TestTrustGateIsActive(t *testing.T) {
	active := WorkTrustGateConfig{Enabled: true, Mode: GovernanceModePerTask}
	require.True(t, active.IsActive())

	off := WorkTrustGateConfig{Enabled: true, Mode: GovernanceModeOff}
	require.False(t, off.IsActive())

	disabled := WorkTrustGateConfig{Enabled: false, Mode: GovernanceModePerTask}
	require.False(t, disabled.IsActive())
}

func TestTrustGateYAMLUnmarshal(t *testing.T) {
	raw := `
enabled: true
mode: per-task
min_score: 80
block_verdicts: [blocked]
warn_verdicts: [risky, acceptable]
warn_is_advisory: false
`
	var g WorkTrustGateConfig
	require.NoError(t, yaml.Unmarshal([]byte(raw), &g))
	applyTrustGateDefaults(&g)
	require.True(t, g.Enabled)
	require.Equal(t, GovernanceModePerTask, g.Mode)
	require.Equal(t, 80.0, g.MinScoreValue())
	require.Equal(t, []string{"risky", "acceptable"}, g.WarnVerdicts)
	require.False(t, g.WarnAdvisory())
}

func TestWorkGatesConfigIncludesTrust(t *testing.T) {
	raw := `
trust:
  enabled: false
  mode: per-task
`
	var gates WorkGatesConfig
	require.NoError(t, yaml.Unmarshal([]byte(raw), &gates))
	applyWorkGatesDefaults(&gates)
	require.False(t, gates.Trust.Enabled)
	require.Equal(t, GovernanceModePerTask, gates.Trust.Mode)
	require.Equal(t, 70.0, gates.Trust.MinScoreValue())
}
