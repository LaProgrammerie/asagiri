package agentobservability_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/agentobservability"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func blockLedgerDir(t *testing.T, repoRoot string) {
	t.Helper()
	path := agentledger.Path(repoRoot)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte{}, 0o444))
}

func TestLedgerFailureBestEffort(t *testing.T) {
	repo := t.TempDir()
	blockLedgerDir(t, repo)
	w := agentobservability.Writer{
		RepoRoot: repo,
		TaskID:   "task-1",
		AgentID:  "dev",
		Mode:     agentobservability.ModeBestEffort,
	}
	err := w.RecordLedger(agentledger.Params{TaskID: "task-1", AgentID: "dev"})
	require.NoError(t, err)
}

func TestLedgerFailureWarnWritesObservabilityFile(t *testing.T) {
	repo := t.TempDir()
	blockLedgerDir(t, repo)
	w := agentobservability.Writer{
		RepoRoot: repo,
		TaskID:   "task-1",
		AgentID:  "dev",
		Mode:     agentobservability.ModeWarn,
	}
	err := w.RecordLedger(agentledger.Params{TaskID: "task-1", AgentID: "dev"})
	require.NoError(t, err)

	warnPath := filepath.Join(agentcontext.AgentLogDir(repo, "task-1", "dev"), "observability.warn")
	require.FileExists(t, warnPath)
	body, readErr := os.ReadFile(warnPath)
	require.NoError(t, readErr)
	require.Contains(t, string(body), "operation=ledger")
}

func TestLedgerFailureStrict(t *testing.T) {
	repo := t.TempDir()
	blockLedgerDir(t, repo)
	w := agentobservability.Writer{
		RepoRoot: repo,
		TaskID:   "task-1",
		AgentID:  "dev",
		Mode:     agentobservability.ModeStrict,
	}
	err := w.RecordLedger(agentledger.Params{TaskID: "task-1", AgentID: "dev"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "ledger")
}

func TestModeFromConfigDefault(t *testing.T) {
	cfg := config.NewTestConfig("repo")
	require.Equal(t, agentobservability.ModeBestEffort, agentobservability.ModeFromConfig(cfg))
	cfg.Work.AgentObservability.Mode = config.AgentObservabilityModeStrict
	require.Equal(t, agentobservability.ModeStrict, agentobservability.ModeFromConfig(cfg))
}
