package agentresolve_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.NewTestConfig("t")
	cfg.Providers = map[string]config.ProviderConfig{
		"exec": {Type: config.ProviderTypeExec, Command: "echo"},
	}
	cfg.Agents = map[string]config.Agent{
		"dev": {Provider: "exec", Command: "echo", Args: []string{"ok"}},
	}
	return cfg
}

func writeDevSpec(t *testing.T, repo string) {
	t.Helper()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := []byte(`id: dev
version: "1.0.0"
role: dev
provider_targets:
  - exec
system_prompt: |
  Agent de test valide.
instructions:
  - instruction a
constraints:
  - constraint a
output_contract:
  format: asagiri-v1
  required_fields:
    - status
metadata:
  labels:
    env: test
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dev.yaml"), body, 0o644))
}

func TestFallbackPreservesLegacyPrompt(t *testing.T) {
	repo := t.TempDir()
	cfg := testConfig(t)
	legacy := "legacy gate prompt exact"
	res, err := agentresolve.Resolve(agentresolve.Params{
		RepoRoot:     repo,
		Config:       cfg,
		Phase:        agentresolve.PhaseGovernance,
		AgentKey:     "dev",
		Feature:      "feat",
		TaskID:       "task-1",
		LegacyPrompt: legacy,
	})
	require.NoError(t, err)
	require.False(t, res.Orchestrated)
	require.Equal(t, legacy, res.Prompt)
}

func TestOrchestratedGovernanceLogs(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo)
	cfg := testConfig(t)
	legacy := "Tu es un validateur de gouvernance post-implémentation."
	res, err := agentresolve.Resolve(agentresolve.Params{
		RepoRoot:     repo,
		Config:       cfg,
		Phase:        agentresolve.PhaseGovernance,
		AgentKey:     "dev",
		Feature:      "feat",
		TaskID:       "task-gov",
		LegacyPrompt: legacy,
	})
	require.NoError(t, err)
	require.True(t, res.Orchestrated)

	dir := agentcontext.AgentLogDirForPhase(repo, "task-gov", "dev", string(agentresolve.PhaseGovernance))
	require.FileExists(t, filepath.Join(dir, "context.json"))
	require.FileExists(t, filepath.Join(dir, "prompt.md"))
	require.FileExists(t, filepath.Join(dir, "invocation.json"))
	require.FileExists(t, filepath.Join(dir, "resolve.json"))
	require.Contains(t, res.Prompt, legacy)
}

func TestDevPhaseUsesHistoricalLogLayout(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo)
	cfg := testConfig(t)
	res, err := agentresolve.Resolve(agentresolve.Params{
		RepoRoot:     repo,
		Config:       cfg,
		Phase:        agentresolve.PhaseDev,
		AgentKey:     "dev",
		Feature:      "feat",
		TaskID:       "task-dev",
		LegacyPrompt: "Implémente la task task-dev",
	})
	require.NoError(t, err)
	require.True(t, res.Orchestrated)
	dir := agentcontext.AgentLogDir(repo, "task-dev", "dev")
	require.FileExists(t, filepath.Join(dir, "prompt.md"))
	_, statErr := os.Stat(filepath.Join(dir, "dev"))
	require.True(t, os.IsNotExist(statErr))
}
