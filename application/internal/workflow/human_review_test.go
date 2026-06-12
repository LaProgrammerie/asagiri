package workflow

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func humanReviewTestService(t *testing.T, humanEnabled, govEnabled bool) (*Service, *sqlite.Store) {
	t.Helper()
	repo := t.TempDir()
	cfg := &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs:   config.Specs{KiroPath: ".kiro/specs"},
		State:   config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:     ".asagiri/worktrees",
			BranchPrefix: "asa",
		},
		Agents: map[string]config.Agent{
			"reviewer": {Command: "echo", Args: []string{"ok"}},
			"dev":      {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				Governance: config.WorkGovernanceGateConfig{
					Enabled: govEnabled,
					Mode:    config.GovernanceModePerTask,
					Agent:   "reviewer",
				},
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: humanEnabled,
					Mode:    config.GovernanceModePerTask,
				},
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

func writeHumanReviewVerdict(t *testing.T, repo, taskID, status string) {
	t.Helper()
	dir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := "human_review:\n  status: " + status + "\n  confidence: 1.0\n  notes:\n    - reviewed\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "human_review.verdict.yaml"), []byte(body), 0o644))
}

func TestHumanReviewDisabledNoOp(t *testing.T) {
	svc, store := humanReviewTestService(t, false, false)
	task := seedTask(t, store, "feat", "task-off", asagiri.StatusImplemented)
	require.NoError(t, svc.processHumanReviewAfterDev(context.Background(), "feat", task))
}

func TestHumanReviewMissingVerdictBlocks(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	task := seedTask(t, store, "feat", "task-miss", asagiri.StatusImplemented)
	err := svc.processHumanReviewAfterDev(context.Background(), "feat", task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "verdict file missing")
	require.Contains(t, err.Error(), "Gate human_review requires action")
	require.Contains(t, err.Error(), "asa gates submit human_review --task task-miss")
	require.Contains(t, err.Error(), "asa continue --yes")
}

func TestHumanReviewSkipsWhenAlreadyPassed(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	task := seedTask(t, store, "feat", "task-skip", asagiri.StatusImplemented)
	writeHumanReviewVerdict(t, svc.repoRoot, "task-skip", "pass")

	require.NoError(t, svc.processHumanReviewAfterDev(context.Background(), "feat", task))
	require.NoError(t, svc.processHumanReviewAfterDev(context.Background(), "feat", task))

	fresh, err := store.GetTask("task-skip")
	require.NoError(t, err)
	var payload struct {
		Gates struct {
			History []struct {
				Gate string `json:"gate"`
			} `json:"history"`
		} `json:"gates"`
	}
	require.NoError(t, json.Unmarshal([]byte(fresh.PayloadJSON), &payload))
	require.Len(t, payload.Gates.History, 1, "second run must not duplicate satisfied human_review")
}

func TestHumanReviewWarnNonAdvisoryBlocks(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs:   config.Specs{KiroPath: ".kiro/specs"},
		State:   config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:     ".asagiri/worktrees",
			BranchPrefix: "asa",
		},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
					WarnIsAdvisory: func() *bool {
						f := false
						return &f
					}(),
				},
			},
		},
	}
	store, err := sqlite.Open(filepath.Join(repo, ".asagiri", "state.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	svc := NewService(repo, cfg, store, false)

	task := seedTask(t, store, "feat", "task-warn-block", asagiri.StatusImplemented)
	writeHumanReviewVerdict(t, repo, "task-warn-block", "warn")

	err = svc.processHumanReviewAfterDev(context.Background(), "feat", task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-advisory")

	fresh, err := store.GetTask("task-warn-block")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
}

func TestHumanReviewPassPersistsHistoryAndLogs(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	task := seedTask(t, store, "feat", "task-pass", asagiri.StatusImplemented)
	writeHumanReviewVerdict(t, svc.repoRoot, "task-pass", "pass")

	require.NoError(t, svc.processHumanReviewAfterDev(context.Background(), "feat", task))

	fresh, err := store.GetTask("task-pass")
	require.NoError(t, err)
	var payload struct {
		Gates struct {
			History []struct {
				Gate   string `json:"gate"`
				Status string `json:"status"`
			} `json:"history"`
		} `json:"gates"`
	}
	require.NoError(t, json.Unmarshal([]byte(fresh.PayloadJSON), &payload))
	require.Len(t, payload.Gates.History, 1)
	require.Equal(t, "human_review", payload.Gates.History[0].Gate)
	require.Equal(t, "pass", payload.Gates.History[0].Status)

	logPath := filepath.Join(svc.repoRoot, ".asagiri", "logs", "task-pass", "gates", "human_review.json")
	body, err := os.ReadFile(logPath)
	require.NoError(t, err)
	var doc gates.LogDocument
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, "pass", doc.Status)
	require.Equal(t, "human_review", doc.GateName)
}

func TestHumanReviewFailMarksTaskFailed(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	task := seedTask(t, store, "feat", "task-fail", asagiri.StatusImplemented)
	writeHumanReviewVerdict(t, svc.repoRoot, "task-fail", "fail")

	err := svc.processHumanReviewAfterDev(context.Background(), "feat", task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "human review gate failed")

	fresh, err := store.GetTask("task-fail")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
}

func TestHumanReviewDryRunPass(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	svc.dryRun = true
	task := seedTask(t, store, "feat", "task-dry", asagiri.StatusImplemented)
	require.NoError(t, svc.processHumanReviewAfterDev(context.Background(), "feat", task))
}

func TestWriteHumanReviewVerdictFile(t *testing.T) {
	repo := t.TempDir()
	path, err := WriteHumanReviewVerdictFile(repo, "task-x", "", "pass", []string{"LGTM"}, "")
	require.NoError(t, err)
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(body), "status: pass")
	require.Contains(t, string(body), "LGTM")
}
