package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func setupResumeTestRepo(t *testing.T) (*Service, string) {
	t.Helper()
	repo := t.TempDir()
	featureDir := filepath.Join(repo, ".kiro", "specs", "agentflow-test")
	require.NoError(t, os.MkdirAll(featureDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "tasks.md"), []byte("- [ ] task A\n"), 0o644))

	cfg := &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs: config.Specs{
			KiroPath:       ".kiro/specs",
			ActiveSpecPath: "docs/ai/active/current-spec.md",
			HandoffPath:    "docs/ai/active/handoff.md",
		},
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:      ".asagiri/worktrees",
			BranchPrefix:  "asa",
			CleanupPolicy: "keep_failed",
		},
		Agents: map[string]config.Agent{
			"cursor": {Command: "cursor", Args: []string{"agent", "run"}},
			"ollama": {Command: "ollama", Args: []string{"run", "qwen2.5-coder:7b"}},
			"codex":  {Command: "codex", Args: []string{"exec"}},
		},
		Work: config.WorkConfig{
			DefaultSpecAgent: "kiro",
			DefaultAgent:     "cursor",
			DefaultReviewer:  "codex",
			DefaultEnricher:  "ollama",
		},
	}

	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	svc := NewService(repo, cfg, store, true)
	runID, _, err := svc.PlanFeature("agentflow-test")
	require.NoError(t, err)
	return svc, runID
}

func TestResumeRunReturnsReportWhenTasksReviewed(t *testing.T) {
	svc, runID := setupResumeTestRepo(t)
	tasks, err := svc.store.ListTasksByFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.NoError(t, svc.store.UpdateTask(&sqlite.Task{
		ID:     tasks[0].ID,
		Status: asagiri.StatusReviewed,
	}))

	next, err := svc.ResumeRun(runID, false)
	require.NoError(t, err)
	require.Equal(t, "report", next)
}

func TestResumeRunExecuteLoopDryRunCompletesPipeline(t *testing.T) {
	svc, runID := setupResumeTestRepo(t)

	steps, err := svc.ResumeRunExecute(context.Background(), runID, false, 0)
	require.NoError(t, err)
	require.Equal(t, []string{"enrich", "dev", "verify", "review", "report"}, steps)

	next, err := svc.ResumeRun(runID, false)
	require.NoError(t, err)
	require.Empty(t, next)
}

func TestResumeRunExecuteLoopRespectsMaxSteps(t *testing.T) {
	svc, runID := setupResumeTestRepo(t)

	steps, err := svc.ResumeRunExecute(context.Background(), runID, false, 1)
	require.ErrorIs(t, err, ErrResumeMaxStepsExceeded)
	require.Equal(t, []string{"enrich"}, steps)

	next, err := svc.ResumeRun(runID, false)
	require.NoError(t, err)
	require.Equal(t, "dev", next)
}

func TestResumeRunDryExecuteRequiresDryRunService(t *testing.T) {
	svc, runID := setupResumeTestRepo(t)
	svc.dryRun = false

	_, err := svc.ResumeRunDryExecute(context.Background(), runID, false, 1)
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrResumeMaxStepsExceeded))
}
