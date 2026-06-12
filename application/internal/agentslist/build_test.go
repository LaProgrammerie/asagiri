package agentslist_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentslist"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func TestBuildEmbeddedDefaultsDeterministic(t *testing.T) {
	repo := t.TempDir()
	report1, err := agentslist.Build(repo, nil)
	require.NoError(t, err)
	report2, err := agentslist.Build(repo, nil)
	require.NoError(t, err)

	require.Equal(t, agentslist.ReportVersion, report1.ReportVersion)
	require.True(t, report1.Registry.UsingEmbeddedDefaults)
	require.NotEmpty(t, report1.Agents)

	body1, err := json.Marshal(report1)
	require.NoError(t, err)
	body2, err := json.Marshal(report2)
	require.NoError(t, err)
	require.JSONEq(t, string(body1), string(body2))

	ids := make([]string, len(report1.Agents))
	for i, a := range report1.Agents {
		ids[i] = a.ID
		require.Equal(t, "embedded", a.Source)
		require.Len(t, a.ContentHash, 64)
		require.NotEmpty(t, a.OutputFormat)
	}
	require.Contains(t, ids, "dev")
}

func TestBuildDiskRegistrySource(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	data := []byte(`id: dev
version: "2.0.0"
role: dev
provider_targets:
  - kiro-cli
system_prompt: |
  Test agent
output_contract:
  format: asagiri-v1
metadata:
  content_hash: "deadbeef"
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dev.yaml"), data, 0o644))

	report, err := agentslist.Build(repo, nil)
	require.NoError(t, err)
	require.False(t, report.Registry.UsingEmbeddedDefaults)
	require.Len(t, report.Agents, 1)
	require.Equal(t, "dev", report.Agents[0].ID)
	require.Equal(t, "disk", report.Agents[0].Source)
	require.Equal(t, "2.0.0", report.Agents[0].Version)
	require.NotEmpty(t, report.Agents[0].Warnings)
}

func TestSemanticHashStableAfterYAMLKeyReorder(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "dev.yaml")

	a := []byte(`id: dev
version: "1.0.0"
role: dev
system_prompt: |
  Same prompt
output_contract:
  format: asagiri-v1
  required_fields:
    - status
    - summary
provider_targets:
  - kiro-cli
  - cursor-cli
`)
	b := []byte(`output_contract:
  required_fields:
    - summary
    - status
  format: asagiri-v1
provider_targets:
  - cursor-cli
  - kiro-cli
role: dev
system_prompt: |
  Same prompt
version: "1.0.0"
id: dev
`)

	require.NoError(t, os.WriteFile(path, a, 0o644))
	reportA, err := agentslist.Build(repo, nil)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, b, 0o644))
	reportB, err := agentslist.Build(repo, nil)
	require.NoError(t, err)

	require.Len(t, reportA.Agents, 1)
	require.Len(t, reportB.Agents, 1)
	require.Equal(t, reportA.Agents[0].ContentHash, reportB.Agents[0].ContentHash)
}

func TestShowUnknownAgent(t *testing.T) {
	repo := t.TempDir()
	_, err := agentslist.Show(repo, "missing", nil)
	require.Error(t, err)
}
