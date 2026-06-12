package workflow

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func verifyReviewGuardService(t *testing.T, humanEnabled, verifyEvidenceEnabled bool) (*Service, *sqlite.Store) {
	t.Helper()
	verify := config.WorkVerifyEvidenceGateConfig{
		Enabled: verifyEvidenceEnabled,
		Mode:    config.GovernanceModePerTask,
		Agent:   "reviewer",
		FailOn:  config.DefaultVerifyEvidenceGateFailOn(),
	}
	return humanReviewTestServiceWithGates(t, humanEnabled, false, verify)
}

func humanReviewTestServiceWithGates(t *testing.T, humanEnabled, govEnabled bool, verify config.WorkVerifyEvidenceGateConfig) (*Service, *sqlite.Store) {
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
				VerifyEvidence: verify,
			},
		},
	}
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return NewService(repo, cfg, store, true), store
}

func TestReviewFeatureBlockedWhenVerifyEvidenceUnsatisfied(t *testing.T) {
	svc, store := verifyReviewGuardService(t, false, true)
	task := seedTask(t, store, "feat", "task-ve-block", asagiri.StatusVerified)

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify evidence gate required before review")
	require.Contains(t, err.Error(), "asa verify feat --task task-ve-block --force")

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
}

func TestReviewFeatureAllowedWhenVerifyEvidenceSatisfied(t *testing.T) {
	svc, store := verifyReviewGuardService(t, false, true)
	task := seedTaskWithGateHistory(t, store, "feat", "task-ve-ok", asagiri.StatusVerified, gates.VerifyEvidenceGateName, "pass")

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusReviewed, fresh.Status)
}

func TestReviewFeatureHumanReviewPendingBeforeVerifyEvidence(t *testing.T) {
	svc, store := verifyReviewGuardService(t, true, true)
	task := seedTask(t, store, "feat", "task-hr-first", asagiri.StatusVerified)

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Gate human_review requires action")
	require.NotContains(t, err.Error(), "verify evidence gate required")

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
}
