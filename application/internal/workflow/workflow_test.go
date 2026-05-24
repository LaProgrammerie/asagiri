package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestPlanFeatureCreatesRunAndTasks(t *testing.T) {
	repo := t.TempDir()
	featureDir := filepath.Join(repo, ".kiro", "specs", "agentflow-test")
	require.NoError(t, os.MkdirAll(featureDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "tasks.md"), []byte("- [ ] task A\n- [ ] task B\n"), 0o644))

	cfg := &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs: config.Specs{
			KiroPath:       ".kiro/specs",
			ActiveSpecPath: "docs/ai/active/current-spec.md",
			HandoffPath:    "docs/ai/active/handoff.md",
		},
		State: config.State{Backend: "sqlite", Path: ".agentflow/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:      ".agentflow/worktrees",
			BranchPrefix:  "agentflow",
			CleanupPolicy: "keep_failed",
		},
		Agents: map[string]config.Agent{
			"cursor": {Command: "cursor", Args: []string{"agent", "run"}},
		},
	}

	dbPath := filepath.Join(repo, ".agentflow", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	svc := NewService(repo, cfg, store, true)
	runID, tasks, err := svc.PlanFeature("agentflow-test")
	require.NoError(t, err)
	require.NotEmpty(t, runID)
	require.Len(t, tasks, 2)

	runs, err := svc.Status(10)
	require.NoError(t, err)
	require.NotEmpty(t, runs)
	require.Equal(t, sqlite.StatusDone, runs[0].Status)
}
