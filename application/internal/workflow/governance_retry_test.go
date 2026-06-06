package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func enableWorktreeDryRun(svc *Service) {
	svc.worktreeMngr.DryRun = true
}

func intPtr(n int) *int {
	return &n
}

func governanceTestServiceMaxRetries(t *testing.T, enabled bool, maxRetries int) (*Service, *sqlite.Store) {
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
				Enabled:    enabled,
				Mode:       config.GovernanceModePerTask,
				Agent:      "reviewer",
				FailOn:     []string{"spec_drift", "architecture_violation", "unexpected_design_change"},
				MaxRetries: intPtr(maxRetries),
			},
		},
	}
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	svc := NewService(repo, cfg, store, false)
	enableWorktreeDryRun(svc)
	return svc, store
}

func failGovernanceYAML() string {
	return `governance:
  status: fail
  confidence: 0.2
  findings:
    - code: spec_drift
      severity: fail
      message: out of scope change
`
}

func passGovernanceYAML() string {
	return `governance:
  status: pass
  confidence: 1
  notes:
    - ok
`
}

func TestGovernanceFailThenPassRetry(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 2)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		if calls == 1 {
			return failGovernanceYAML(), nil
		}
		return passGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-retry-pass", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-retry-pass", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	require.NoError(t, svc.devTaskWithGovernanceRetries(context.Background(), run, "feat", task, a, true))
	require.Equal(t, 2, calls)

	fresh, err := store.GetTask("task-retry-pass")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 1, canonical.Governance.Retries)
	require.Len(t, canonical.Governance.History, 2)
	require.Equal(t, 0, canonical.Governance.History[0].Retry)
	require.Equal(t, "fail", canonical.Governance.History[0].Status)
	require.Equal(t, 1, canonical.Governance.History[1].Retry)
	require.Equal(t, "pass", canonical.Governance.History[1].Status)
}

func TestGovernanceFailExhaustedMaxRetries(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 2)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		return failGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-exhaust", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-exhaust", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	err = svc.devTaskWithGovernanceRetries(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "governance gate failed after 2 retries (max 2)")
	require.Equal(t, 3, calls)

	fresh, err := store.GetTask("task-exhaust")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 2, canonical.Governance.Retries)
	require.Len(t, canonical.Governance.History, 3)
	require.Equal(t, 0, canonical.Governance.History[0].Retry)
	require.Equal(t, 1, canonical.Governance.History[1].Retry)
	require.Equal(t, 2, canonical.Governance.History[2].Retry)
}

func TestGovernanceMaxRetriesZeroBlocksImmediately(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 0)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		return failGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-max0", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-max0", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	err = svc.devTaskWithGovernanceRetries(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "governance gate failed after 0 retries (max 0)")
	require.Equal(t, 1, calls)

	fresh, err := store.GetTask("task-max0")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 0, canonical.Governance.Retries)
	require.Len(t, canonical.Governance.History, 1)
	require.Equal(t, 0, canonical.Governance.History[0].Retry)
}

func TestGovernanceMaxRetriesOneRequiresTwoFails(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 1)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		return failGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-max1", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-max1", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	err = svc.devTaskWithGovernanceRetries(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "governance gate failed after 1 retries (max 1)")
	require.Equal(t, 2, calls)

	fresh, err := store.GetTask("task-max1")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 1, canonical.Governance.Retries)
	require.Len(t, canonical.Governance.History, 2)
	require.Equal(t, 0, canonical.Governance.History[0].Retry)
	require.Equal(t, 1, canonical.Governance.History[1].Retry)
}

func TestGovernanceParseErrorCountsAsRetryableFail(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 2)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		if calls == 1 {
			return "not valid governance output", nil
		}
		return passGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-parse", asagiri.StatusImplemented)
	task.WorktreePath = filepath.Join(svc.repoRoot, "wt")
	require.NoError(t, os.MkdirAll(task.WorktreePath, 0o755))
	require.NoError(t, store.UpdateTask(&sqlite.Task{ID: task.ID, WorktreePath: task.WorktreePath}))

	outcome, next, err := svc.processGovernanceAfterDev(context.Background(), "feat", task)
	require.NoError(t, err)
	require.Equal(t, governanceRetryDev, outcome)

	outcome, _, err = svc.processGovernanceAfterDev(context.Background(), "feat", next)
	require.NoError(t, err)
	require.Equal(t, governanceOK, outcome)
	require.Equal(t, 2, calls)

	fresh, err := store.GetTask("task-parse")
	require.NoError(t, err)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 1, canonical.Governance.Retries)
	require.Equal(t, "fail", canonical.Governance.History[0].Status)
	require.NotEmpty(t, canonical.Governance.History[0].ParseError)
}

func TestGovernanceAgentErrorCountsAsRetryableFail(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 2)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		if calls == 1 {
			return "", context.Canceled
		}
		return passGovernanceYAML(), nil
	}

	task := seedTask(t, store, "feat", "task-agent-err", asagiri.StatusImplemented)

	outcome, next, err := svc.processGovernanceAfterDev(context.Background(), "feat", task)
	require.NoError(t, err)
	require.Equal(t, governanceRetryDev, outcome)

	outcome, _, err = svc.processGovernanceAfterDev(context.Background(), "feat", next)
	require.NoError(t, err)
	require.Equal(t, governanceOK, outcome)

	fresh, err := store.GetTask("task-agent-err")
	require.NoError(t, err)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 1, canonical.Governance.Retries)
	require.Equal(t, "fail", canonical.Governance.History[0].Status)
}

func TestGovernanceWarnDoesNotRetry(t *testing.T) {
	svc, store := governanceTestServiceMaxRetries(t, true, 2)
	calls := 0
	svc.governanceAgentHook = func(_ context.Context, _ string) (string, error) {
		calls++
		return `governance:
  status: warn
  confidence: 0.6
  notes:
    - advisory
`, nil
	}

	task := seedTask(t, store, "feat", "task-warn-no-retry", asagiri.StatusImplemented)
	outcome, _, err := svc.processGovernanceAfterDev(context.Background(), "feat", task)
	require.NoError(t, err)
	require.Equal(t, governanceOK, outcome)
	require.Equal(t, 1, calls)

	fresh, err := store.GetTask("task-warn-no-retry")
	require.NoError(t, err)
	canonical, err := payloadToCanonical(fresh.PayloadJSON)
	require.NoError(t, err)
	require.Equal(t, 0, canonical.Governance.Retries)
	require.Len(t, canonical.Governance.History, 1)
	require.Equal(t, "warn", canonical.Governance.History[0].Status)
}

func TestDevOneTaskReusesWorktree(t *testing.T) {
	svc, store := governanceTestService(t, true)
	enableWorktreeDryRun(svc)
	existing := filepath.Join(svc.repoRoot, ".asagiri", "worktrees", "feat", "task-wt")
	require.NoError(t, os.MkdirAll(existing, 0o755))

	task := seedTask(t, store, "feat", "task-wt", asagiri.StatusRunning)
	require.NoError(t, store.UpdateTask(&sqlite.Task{ID: task.ID, WorktreePath: existing}))
	task.WorktreePath = existing

	run := &sqlite.Run{ID: "run-wt", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	updated, err := svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)
	require.Equal(t, existing, updated.WorktreePath)

	fresh, err := store.GetTask("task-wt")
	require.NoError(t, err)
	require.Equal(t, existing, fresh.WorktreePath)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}
