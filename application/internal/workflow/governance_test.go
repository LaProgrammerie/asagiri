package workflow

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func governanceTestService(t *testing.T, enabled bool) (*Service, *sqlite.Store) {
	t.Helper()
	repo := t.TempDir()
	cfg := &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs: config.Specs{
			KiroPath: ".kiro/specs",
		},
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:     ".asagiri/worktrees",
			BranchPrefix: "asa",
		},
		Agents: map[string]config.Agent{
			"reviewer": {Command: "echo", Args: []string{"ok"}},
			"dev":      {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			Governance: config.WorkGovernanceConfig{
				Enabled: enabled,
				Mode:    config.GovernanceModePerTask,
				Agent:   "reviewer",
				FailOn:  []string{"spec_drift", "architecture_violation", "unexpected_design_change"},
			},
		},
	}
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	svc := NewService(repo, cfg, store, false)
	return svc, store
}

func seedTask(t *testing.T, store *sqlite.Store, feature, id, status string) sqlite.Task {
	t.Helper()
	task := asagiri.Task{
		ID:      id,
		Title:   "test",
		Feature: feature,
		Status:  status,
	}
	payload, err := canonicalToPayload(task)
	require.NoError(t, err)
	row := &sqlite.Task{
		ID:          id,
		RunID:       "run-test",
		Feature:     feature,
		Status:      status,
		PayloadJSON: payload,
	}
	require.NoError(t, store.CreateTask(row))
	return *row
}

func TestGovernanceDryRunSimulatedPass(t *testing.T) {
	svc, store := governanceTestService(t, true)
	svc.dryRun = true
	task := seedTask(t, store, "feat", "task-1", asagiri.StatusImplemented)

	err := svc.applyGovernanceAfterDev(context.Background(), "feat", task, "")
	require.NoError(t, err)

	logPath := gateLogJSONPath(svc.repoRoot, "task-1", "governance")
	require.FileExists(t, logPath)
	body, err := os.ReadFile(logPath)
	require.NoError(t, err)
	var doc gateLogDocument
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, "pass", doc.Status)
	require.True(t, doc.DryRun)
}

func TestGovernanceWarnDoesNotBlock(t *testing.T) {
	svc, store := governanceTestService(t, true)
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		return `governance:
  status: warn
  confidence: 0.6
  notes:
    - advisory only
`, nil
	}
	task := seedTask(t, store, "feat", "task-warn", asagiri.StatusImplemented)

	err := svc.applyGovernanceAfterDev(context.Background(), "feat", task, "")
	require.NoError(t, err)

	fresh, err := store.GetTask("task-warn")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}

func TestGovernanceFailBlocksWhenRetriesExhausted(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 1)
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		return failGovernanceYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-fail", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-fail", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	err = svc.devTaskWithGovernanceRetries(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "governance gate failed after 1 retries (max 1)")

	fresh, err := store.GetTask("task-fail")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 1, canonical.Governance.Retries)
	require.Len(t, canonical.Governance.History, 2)
}

func TestGovernanceDisabledNoAgentCall(t *testing.T) {
	svc, store := governanceTestService(t, false)
	called := false
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		called = true
		return "", nil
	}
	task := seedTask(t, store, "feat", "task-off", asagiri.StatusImplemented)

	err := svc.applyGovernanceAfterDev(context.Background(), "feat", task, "")
	require.NoError(t, err)
	require.False(t, called)

	logPath := gateLogJSONPath(svc.repoRoot, "task-off", "governance")
	require.NoFileExists(t, logPath)
}

func TestDevTaskWithGovernanceDisabledUnchanged(t *testing.T) {
	svc, store := governanceTestService(t, false)
	enableWorktreeDryRun(svc)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		return failGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-dev-off", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-dev-off", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	require.NoError(t, svc.devTaskWithGovernanceRetries(context.Background(), run, "feat", task, a, true))
	require.Equal(t, 0, calls)

	fresh, err := store.GetTask("task-dev-off")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Nil(t, canonical.Governance)
}

func TestGovernanceLogsWritten(t *testing.T) {
	svc, store := governanceTestService(t, true)
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		return `governance:
  status: pass
  confidence: 1
  notes:
    - validated
`, nil
	}
	task := seedTask(t, store, "feat", "task-log", asagiri.StatusImplemented)

	require.NoError(t, svc.applyGovernanceAfterDev(context.Background(), "feat", task, ""))

	logPath := gateLogJSONPath(svc.repoRoot, "task-log", "governance")
	require.FileExists(t, logPath)
	require.FileExists(t, gateLogMarkdownPath(svc.repoRoot, "task-log", "governance"))

	fresh, err := store.GetTask("task-log")
	require.NoError(t, err)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.NotNil(t, canonical.Governance)
	require.Len(t, canonical.Governance.History, 1)
	require.Equal(t, "pass", canonical.Governance.History[0].Status)
}
