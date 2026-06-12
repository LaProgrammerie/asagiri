package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func blockLedgerAppend(t *testing.T, repoRoot string) {
	t.Helper()
	path := agentledger.Path(repoRoot)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte{}, 0o444))
}

func setAgentObservabilityMode(svc *Service, mode string) {
	svc.cfg.Work.AgentObservability.Mode = mode
	config.NormalizeWorkAgentObservability(&svc.cfg.Work.AgentObservability)
}

func TestDevLedgerFailureBestEffort(t *testing.T) {
	svc, store := devOrchestrateService(t, true)
	setAgentObservabilityMode(svc, config.AgentObservabilityModeBestEffort)
	blockLedgerAppend(t, svc.repoRoot)

	task := seedTask(t, store, "feat", "task-led-be", asagiri.StatusEnriched)
	run := &sqlite.Run{ID: "run-led-be", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	report, err := agentledger.List(svc.repoRoot, agentledger.ListOptions{TaskID: "task-led-be"})
	require.NoError(t, err)
	require.Empty(t, report.Entries)
}

func TestDevLedgerFailureWarn(t *testing.T) {
	svc, store := devOrchestrateService(t, true)
	setAgentObservabilityMode(svc, config.AgentObservabilityModeWarn)
	blockLedgerAppend(t, svc.repoRoot)

	task := seedTask(t, store, "feat", "task-led-warn", asagiri.StatusEnriched)
	run := &sqlite.Run{ID: "run-led-warn", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	warnPath := filepath.Join(agentcontext.AgentLogDir(svc.repoRoot, "task-led-warn", "dev"), "observability.warn")
	require.FileExists(t, warnPath)
	body, readErr := os.ReadFile(warnPath)
	require.NoError(t, readErr)
	require.Contains(t, string(body), "operation=ledger")

	report, err := agentledger.List(svc.repoRoot, agentledger.ListOptions{TaskID: "task-led-warn"})
	require.NoError(t, err)
	require.Empty(t, report.Entries)
}

func TestDevLedgerFailureStrict(t *testing.T) {
	svc, store := devOrchestrateService(t, true)
	setAgentObservabilityMode(svc, config.AgentObservabilityModeStrict)
	blockLedgerAppend(t, svc.repoRoot)

	task := seedTask(t, store, "feat", "task-led-strict", asagiri.StatusEnriched)
	run := &sqlite.Run{ID: "run-led-strict", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ledger")

	report, err := agentledger.List(svc.repoRoot, agentledger.ListOptions{TaskID: "task-led-strict"})
	require.NoError(t, err)
	require.Empty(t, report.Entries)
}
