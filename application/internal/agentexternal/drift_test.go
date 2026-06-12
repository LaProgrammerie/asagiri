package agentexternal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentexternal"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestExternalDriftDetectsHashMismatch(t *testing.T) {
	dir := t.TempDir()
	extFile := filepath.Join(dir, "dev.md")
	require.NoError(t, os.WriteFile(extFile, []byte("# manual profile\n"), 0o644))

	spec, err := agentspec.Parse([]byte(`id: dev
version: "1.0.0"
role: dev
system_prompt: inline
output_contract:
  format: asagiri-v1
external:
  kind: cursor-agent
  path: `+extFile+`
  last_synced_hash: "0000000000000000000000000000000000000000000000000000000000000000"
`), "dev.yaml")
	require.NoError(t, err)

	drift, ok := agentexternal.ExternalDrift(spec, nil, "dev")
	require.True(t, ok)
	require.Equal(t, "external_drift", drift.Kind)
	require.Contains(t, drift.FixCLI, "asa agents external sync --write --agent dev")
}

func TestExternalDriftDetectsMissingFile(t *testing.T) {
	dir := t.TempDir()
	extFile := filepath.Join(dir, "missing.md")

	spec, err := agentspec.Parse([]byte(`id: dev
version: "1.0.0"
role: dev
system_prompt: inline
output_contract:
  format: asagiri-v1
external:
  kind: kiro-agent
  path: `+extFile+`
`), "dev.yaml")
	require.NoError(t, err)

	drift, ok := agentexternal.ExternalDrift(spec, nil, "dev")
	require.True(t, ok)
	require.Equal(t, "external_missing", drift.Kind)
	require.Contains(t, drift.Message, "absent")
}

func TestExternalDriftIgnoresUnconfiguredPath(t *testing.T) {
	spec, err := agentspec.Parse([]byte(`id: dev
version: "1.0.0"
role: dev
system_prompt: inline
output_contract:
  format: asagiri-v1
`), "dev.yaml")
	require.NoError(t, err)

	cfg := &config.Config{
		Agents: map[string]config.Agent{
			"dev": {Provider: "kiro-cli"},
		},
	}
	_, ok := agentexternal.ExternalDrift(spec, cfg, "dev")
	require.False(t, ok)
}
