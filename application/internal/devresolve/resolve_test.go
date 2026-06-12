package devresolve_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/devresolve"
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

func writeDevSpec(t *testing.T, repo string, filename string, body []byte) {
	t.Helper()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), body, 0o644))
}

func baseParams(repo string, cfg *config.Config) devresolve.Params {
	return devresolve.Params{
		RepoRoot:     repo,
		Config:       cfg,
		AgentKey:     "dev",
		RunID:        "run-snap",
		Feature:      "feat-orchestrate",
		TaskID:       "task-snap",
		ContextFiles: []string{"docs/ai/active/handoff.md"},
	}
}

func TestLegacyDevPromptExact(t *testing.T) {
	require.Equal(t, "Implémente la task task-42", devresolve.LegacyDevPrompt("task-42"))
}

func TestFallbackWhenAgentSpecMissing(t *testing.T) {
	repo := t.TempDir()
	cfg := testConfig(t)
	res, err := devresolve.Resolve(baseParams(repo, cfg))
	require.NoError(t, err)
	require.False(t, res.Orchestrated)
	require.Equal(t, devresolve.LegacyDevPrompt("task-snap"), res.Prompt)
	require.NotEmpty(t, res.Warning)

	dir := agentcontext.AgentLogDir(repo, "task-snap", "dev")
	require.FileExists(t, filepath.Join(dir, "resolve.json"))
	require.NoFileExists(t, filepath.Join(dir, "prompt.md"))
}

func TestFallbackWhenAgentSpecInvalid(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo, "dev.yaml", []byte("id: [broken"))
	cfg := testConfig(t)
	res, err := devresolve.Resolve(baseParams(repo, cfg))
	require.NoError(t, err)
	require.False(t, res.Orchestrated)
	require.Equal(t, devresolve.LegacyDevPrompt("task-snap"), res.Prompt)
	require.Contains(t, res.Warning, "parse YAML")
}

func orchestratedDevSpecYAML() []byte {
	return []byte(`id: dev
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
}

func TestOrchestratedPromptWithValidSpec(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo, "dev.yaml", orchestratedDevSpecYAML())

	cfg := testConfig(t)
	res, err := devresolve.Resolve(baseParams(repo, cfg))
	require.NoError(t, err)
	require.True(t, res.Orchestrated)
	require.Contains(t, res.Prompt, "Tu es exécuté par Asagiri en mode orchestré")
	require.Contains(t, res.Prompt, "## Task prompt")
	require.Contains(t, res.Prompt, devresolve.LegacyDevPrompt("task-snap"))
	require.NotEqual(t, devresolve.LegacyDevPrompt("task-snap"), res.Prompt)
}

func TestOrchestrationLogsWritten(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo, "dev.yaml", orchestratedDevSpecYAML())

	cfg := testConfig(t)
	res, err := devresolve.Resolve(baseParams(repo, cfg))
	require.NoError(t, err)
	require.True(t, res.Orchestrated)

	dir := agentcontext.AgentLogDir(repo, "task-snap", "dev")
	require.FileExists(t, filepath.Join(dir, "context.json"))
	require.FileExists(t, filepath.Join(dir, "prompt.md"))
	require.FileExists(t, filepath.Join(dir, "invocation.json"))
	require.FileExists(t, filepath.Join(dir, "resolve.json"))

	var inv agentadapter.RenderedInvocation
	require.NoError(t, json.Unmarshal(mustRead(t, filepath.Join(dir, "invocation.json")), &inv))
	require.Equal(t, res.Prompt, inv.StdinPrompt)
	require.Equal(t, "echo", inv.Command)
}

func TestAdapterRenderOnlyNoSubprocess(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo, "dev.yaml", orchestratedDevSpecYAML())

	cfg := testConfig(t)
	_, err := devresolve.Resolve(baseParams(repo, cfg))
	require.NoError(t, err)

	dir := agentcontext.AgentLogDir(repo, "task-snap", "dev")
	invPath := filepath.Join(dir, "invocation.json")
	require.FileExists(t, invPath)
	body := mustRead(t, invPath)
	require.NotContains(t, string(body), "subprocess")
}

func TestResolvePromptSnapshot(t *testing.T) {
	repo := t.TempDir()
	specBody := []byte(`id: dev
version: "1.0.0"
role: dev
provider_targets:
  - exec
system_prompt: |
  Agent snapshot fixe.
instructions:
  - instruction fixe
constraints:
  - constraint fixe
output_contract:
  format: asagiri-v1
  required_fields:
    - status
metadata:
  labels:
    env: snapshot
`)
	writeDevSpec(t, repo, "dev.yaml", specBody)
	cfg := testConfig(t)
	params := baseParams(repo, cfg)
	params.ContextFiles = []string{"z-context.md", "a-context.md"}

	res, err := devresolve.Resolve(params)
	require.NoError(t, err)
	require.True(t, res.Orchestrated)

	goldenPath := filepath.Join("testdata", "prompt.golden")
	golden, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
		require.NoError(t, os.WriteFile(goldenPath, []byte(res.Prompt), 0o644))
		t.Fatalf("golden créé — relancer le test")
	}
	require.NoError(t, err)
	require.Equal(t, string(golden), res.Prompt)
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}
