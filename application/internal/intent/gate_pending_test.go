package intent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func snapWithPendingGate(feature, taskID, status string, pg *gates.PendingGate) StateSnapshot {
	return StateSnapshot{
		ActiveFeature: feature,
		Features: []FeatureState{{
			Name: feature, NextTaskID: taskID, NextTaskStatus: status, PendingGate: pg,
		}},
	}
}

func TestRecommendNextHumanReviewSubmitPending(t *testing.T) {
	pg := &gates.PendingGate{
		Gate: gates.HumanReviewGateName, Scope: "task-hr", Blocking: true, Phase: gates.PendingPhaseSubmit,
	}
	snap := snapWithPendingGate("feat", "task-hr", asagiri.StatusImplemented, pg)
	rec, err := RecommendNext(snap, "feat", &config.Config{})
	require.NoError(t, err)
	require.Equal(t, "human_review", rec.Action)
	require.Contains(t, rec.Primitive, "gates submit human_review")
	require.Contains(t, rec.Reason, "Gate human_review requires action")
	require.NotContains(t, rec.Primitive, "verify")
}

func TestRecommendNextHumanReviewResumePending(t *testing.T) {
	pg := &gates.PendingGate{
		Gate: gates.HumanReviewGateName, Scope: "task-hr", Blocking: true, Phase: gates.PendingPhaseResume,
	}
	snap := snapWithPendingGate("feat", "task-hr", asagiri.StatusImplemented, pg)
	rec, err := RecommendNext(snap, "feat", &config.Config{})
	require.NoError(t, err)
	require.Equal(t, "human_review", rec.Action)
	require.Equal(t, "asa continue --yes", rec.Primitive)
}

func TestRecommendNextImplementedWithoutPendingGate(t *testing.T) {
	snap := snapWithPendingGate("feat", "task-1", asagiri.StatusImplemented, nil)
	rec, err := RecommendNext(snap, "feat", nil)
	require.NoError(t, err)
	require.Equal(t, "verify", rec.Action)
}

func TestEvaluateConditionBlocksVerifyWhenGatePending(t *testing.T) {
	pg := &gates.PendingGate{Gate: gates.HumanReviewGateName, Scope: "t1", Blocking: true, Phase: gates.PendingPhaseSubmit}
	fs := FeatureState{NextTaskStatus: asagiri.StatusImplemented, PendingGate: pg}
	require.False(t, EvaluateCondition("implementation_done", ResolvedIntent{}, fs, WorkOptions{}))
	require.True(t, EvaluateCondition("gate_blocking", ResolvedIntent{}, fs, WorkOptions{}))
}

func TestPlannerAddsDevStepForGateResume(t *testing.T) {
	pg := &gates.PendingGate{
		Gate: gates.HumanReviewGateName, Scope: "task-hr", Blocking: true, Phase: gates.PendingPhaseResume,
	}
	snap := snapWithPendingGate("feat", "task-hr", asagiri.StatusImplemented, pg)
	cfg := &config.Config{Work: config.WorkConfig{AutoVerify: true, DefaultAgent: "dev"}}
	plan, err := (&DefaultPlanner{}).BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentResume, Feature: "feat", TaskID: "task-hr",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)
	fs := FeatureStateFor(snap, "feat")
	var commands []string
	for _, s := range plan.Steps {
		if EvaluateCondition(s.Condition, plan.Intent, fs, WorkOptions{}) {
			commands = append(commands, s.Command)
		}
	}
	require.Contains(t, commands, "dev")
	require.NotContains(t, commands, "verify")
}

func TestBuildSnapshotAttachesPendingGate(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		Specs: config.Specs{KiroPath: ".kiro/specs"},
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	run := &sqlite.Run{ID: "run-1", Feature: "feat", Status: sqlite.StatusRunning}
	require.NoError(t, store.CreateRun(run))
	payload := `{"title":"t","feature":"feat","status":"implemented"}`
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-snap", RunID: run.ID, Feature: "feat", Status: asagiri.StatusImplemented, PayloadJSON: payload,
	}))

	snap, err := BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	var fs FeatureState
	for _, f := range snap.Features {
		if f.Name == "feat" {
			fs = f
			break
		}
	}
	require.NotNil(t, fs.PendingGate)
	require.Equal(t, gates.PendingPhaseSubmit, fs.PendingGate.Phase)
}

func TestWorkPlanSkipsVerifyWhenGateSubmitPending(t *testing.T) {
	pg := &gates.PendingGate{
		Gate: gates.HumanReviewGateName, Scope: "task-hr", Blocking: true, Phase: gates.PendingPhaseSubmit,
	}
	snap := snapWithPendingGate("feat", "task-hr", asagiri.StatusImplemented, pg)
	cfg := &config.Config{Work: config.WorkConfig{AutoVerify: true, DefaultAgent: "dev"}}
	plan, err := (&DefaultPlanner{}).BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "feat", TaskID: "task-hr",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)
	fs := FeatureStateFor(snap, "feat")
	var commands []string
	for _, s := range plan.Steps {
		if EvaluateCondition(s.Condition, plan.Intent, fs, WorkOptions{}) {
			commands = append(commands, s.Command)
		}
	}
	require.NotContains(t, commands, "verify")
	require.NotContains(t, commands, "dev")
}

func TestRecommendNextAfterHumanReviewPassInSnapshot(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
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

	run := &sqlite.Run{ID: "run-pass", Feature: "feat", Status: sqlite.StatusRunning}
	require.NoError(t, store.CreateRun(run))
	payload := `{"gates":{"history":[{"gate":"human_review","status":"pass","confidence":1}]}}`
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-pass", RunID: run.ID, Feature: "feat", Status: asagiri.StatusImplemented, PayloadJSON: payload,
	}))

	snap, err := BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	fs := FeatureStateFor(snap, "feat")
	require.Nil(t, fs.PendingGate)

	rec, err := RecommendNext(snap, "feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "verify", rec.Action)
}

func TestPlannerNoGateImpactWhenDisabled(t *testing.T) {
	snap := snapWithPendingGate("feat", "task-1", asagiri.StatusEnriched, nil)
	cfg := &config.Config{Work: config.WorkConfig{DefaultAgent: "dev"}}
	plan, err := (&DefaultPlanner{}).BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "feat", TaskID: "task-1",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)
	commands := make([]string, len(plan.Steps))
	for i, s := range plan.Steps {
		commands[i] = s.Command
	}
	require.Contains(t, commands, "dev")
}
