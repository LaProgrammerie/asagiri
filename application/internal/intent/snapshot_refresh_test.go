package intent

import (
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

func TestRefreshFeatureTaskStateRecomputesGateFlags(t *testing.T) {
	repo, store, cfg := refreshTestFixture(t, true, false)

	const taskID = "task-flags"
	payload, err := json.Marshal(map[string]any{
		"gates": map[string]any{
			"history": []map[string]any{
				{"gate": "enrich", "status": "pass", "confidence": 0.9},
			},
		},
	})
	require.NoError(t, err)
	seedRefreshTask(t, store, taskID, asagiri.StatusEnriched, string(payload))

	stale := FeatureState{
		Name: "feat", NextTaskID: taskID, NextTaskStatus: asagiri.StatusPlanned, EnrichGateBlocksDev: true,
	}
	refreshed := RefreshFeatureTaskState(repo, cfg, store, "feat", stale)
	require.Equal(t, asagiri.StatusEnriched, refreshed.NextTaskStatus)
	require.False(t, refreshed.EnrichGateBlocksDev)
}

func TestRefreshFeatureTaskStatePreservesSubmitPending(t *testing.T) {
	repo, store, cfg := refreshTestFixture(t, false, true)

	const taskID = "task-hr-submit"
	seedRefreshTask(t, store, taskID, asagiri.StatusImplemented, "")

	snap, err := BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	fs := FeatureStateFor(snap, "feat")
	require.Equal(t, gates.PendingPhaseSubmit, fs.PendingGate.Phase)

	refreshed := RefreshFeatureTaskState(repo, cfg, store, "feat", fs)
	require.Equal(t, gates.PendingPhaseSubmit, refreshed.PendingGate.Phase)
}

func refreshTestFixture(t *testing.T, enrich, humanReview bool) (string, *sqlite.Store, *config.Config) {
	t.Helper()
	repo := t.TempDir()
	cfg := &config.Config{
		Specs: config.Specs{KiroPath: ".kiro/specs"},
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{
					Enabled: enrich,
					Mode:    config.GovernanceModePerTask,
				},
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: humanReview,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))
	store, err := sqlite.Open(filepath.Join(repo, ".asagiri", "state.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return repo, store, cfg
}

func seedRefreshTask(t *testing.T, store *sqlite.Store, id, status, payload string) {
	t.Helper()
	run := &sqlite.Run{ID: "run-refresh", Feature: "feat", Status: sqlite.StatusRunning}
	require.NoError(t, store.CreateRun(run))
	if payload == "" {
		task := asagiri.Task{ID: id, Title: "t", Feature: "feat", Status: status}
		b, err := json.Marshal(task)
		require.NoError(t, err)
		payload = string(b)
	}
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: id, RunID: run.ID, Feature: "feat", Status: status, PayloadJSON: payload,
	}))
}
