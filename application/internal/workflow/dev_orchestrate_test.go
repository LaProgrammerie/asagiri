package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/devresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func devOrchestrateService(t *testing.T, withAgentSpec bool) (*Service, *sqlite.Store) {
	t.Helper()
	svc, store := devEnrichGateService(t, false)
	if withAgentSpec {
		dir := filepath.Join(svc.repoRoot, agentspec.RegistryDir)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		fixed := []byte(`id: dev
version: "1.0.0"
role: dev
provider_targets:
  - exec
system_prompt: |
  Agent de test workflow.
instructions:
  - instruction workflow
constraints:
  - constraint workflow
output_contract:
  format: asagiri-v1
  required_fields:
    - status
metadata:
  labels:
    env: workflow-test
`)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "dev.yaml"), fixed, 0o644))
	}
	return svc, store
}

func TestDevOneTaskOrchestrationLogsWithAgentSpec(t *testing.T) {
	svc, store := devOrchestrateService(t, true)
	task := seedTask(t, store, "feat", "task-orch", asagiri.StatusEnriched)
	run := &sqlite.Run{ID: "run-orch", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	dir := agentcontext.AgentLogDir(svc.repoRoot, "task-orch", "dev")
	require.FileExists(t, filepath.Join(dir, "prompt.md"))
	require.FileExists(t, filepath.Join(dir, "context.json"))
	require.FileExists(t, filepath.Join(dir, "invocation.json"))

	body, err := os.ReadFile(filepath.Join(dir, "prompt.md"))
	require.NoError(t, err)
	require.Contains(t, string(body), "Tu es exécuté par Asagiri en mode orchestré")
	require.Contains(t, string(body), devresolve.LegacyDevPrompt("task-orch"))
	require.FileExists(t, filepath.Join(dir, "contract.json"))

	report, err := agentledger.List(svc.repoRoot, agentledger.ListOptions{TaskID: "task-orch"})
	require.NoError(t, err)
	require.Len(t, report.Entries, 1)
	require.Equal(t, "run-orch", report.Entries[0].RunID)
	require.Equal(t, "dev", report.Entries[0].AgentID)
	require.NotEmpty(t, report.Entries[0].PromptHash)
	require.NotEmpty(t, report.Entries[0].ContextHash)
	require.NotNil(t, report.Entries[0].ContractValid)
}

func TestDevOneTaskNoContractJSONOnFallback(t *testing.T) {
	svc, store := devOrchestrateService(t, false)
	task := seedTask(t, store, "feat", "task-fallback", asagiri.StatusEnriched)
	run := &sqlite.Run{ID: "run-fallback", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	dir := agentcontext.AgentLogDir(svc.repoRoot, "task-fallback", "dev")
	require.FileExists(t, filepath.Join(dir, "resolve.json"))
	require.NoFileExists(t, filepath.Join(dir, "prompt.md"))

	warnPath := filepath.Join(svc.repoRoot, ".asagiri", "logs", "task-fallback", "dev-orchestration.warn")
	require.FileExists(t, warnPath)
	require.NoFileExists(t, filepath.Join(dir, "contract.json"))
}
