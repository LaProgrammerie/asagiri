package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestEvaluateConditions(t *testing.T) {
	fs := FeatureState{HasLocalSpec: false, HasTasks: false, NextTaskStatus: asagiri.StatusPlanned}
	intent := ResolvedIntent{RequiresSync: true, Source: "notion"}
	opts := WorkOptions{}

	require.True(t, EvaluateCondition("source_requires_sync", intent, fs, opts))
	require.True(t, EvaluateCondition("no_local_spec", intent, fs, opts))
	require.True(t, EvaluateCondition("no_tasks", intent, fs, opts))
	require.True(t, EvaluateCondition("task_not_enriched", intent, fs, opts))

	fs.NextTaskStatus = asagiri.StatusImplemented
	require.True(t, EvaluateCondition("implementation_done", intent, fs, opts))

	fs.NextTaskStatus = asagiri.StatusVerifyFailed
	require.True(t, EvaluateCondition("verification_failed", intent, fs, opts))

	opts.NoReview = true
	require.False(t, EvaluateCondition("review_enabled", intent, fs, opts))
}

func TestPlannerDevelopSteps(t *testing.T) {
	p := &DefaultPlanner{}
	cfg := &config.Config{Work: config.WorkConfig{AutoVerify: true, DefaultAgent: "cursor"}}
	snap := StateSnapshot{Features: []FeatureState{
		{Name: "billing-v2", HasLocalSpec: true, HasTasks: true, NextTaskID: "task-003", NextTaskStatus: asagiri.StatusPlanned},
	}}
	plan, err := p.BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "billing-v2", TaskID: "task-003",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, plan.Steps)
	commands := make([]string, len(plan.Steps))
	for i, s := range plan.Steps {
		commands[i] = s.Command
	}
	require.Contains(t, commands, "enrich")
	require.Contains(t, commands, "dev")
	require.Contains(t, commands, "verify")
}

func TestEvaluateConditionUnknownReturnsFalse(t *testing.T) {
	require.False(t, EvaluateCondition("nonexistent_condition_xyz", ResolvedIntent{}, FeatureState{}, WorkOptions{}))
}

func TestRecommendNextReadsAgentFromCfg(t *testing.T) {
	cfg := &config.Config{Work: config.WorkConfig{DefaultAgent: "claude-code", DefaultReviewer: "kiro-review"}}
	snap := StateSnapshot{
		ActiveFeature: "my-feat",
		Features: []FeatureState{{
			Name: "my-feat", NextTaskID: "t1", NextTaskStatus: "enriched",
		}},
	}
	rec, err := RecommendNext(snap, "my-feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "dev", rec.Action)
	require.Contains(t, rec.Primitive, "--agent claude-code")
}

func TestRecommendNextReviewerFromCfg(t *testing.T) {
	cfg := &config.Config{Work: config.WorkConfig{DefaultReviewer: "my-reviewer"}}
	snap := StateSnapshot{
		ActiveFeature: "f",
		Features: []FeatureState{{Name: "f", NextTaskID: "t1", NextTaskStatus: "verified"}},
	}
	rec, err := RecommendNext(snap, "f", cfg)
	require.NoError(t, err)
	require.Equal(t, "review", rec.Action)
	require.Contains(t, rec.Primitive, "--agent my-reviewer")
}

func TestRecommendNextNilCfgFallback(t *testing.T) {
	snap := StateSnapshot{
		ActiveFeature: "f",
		Features: []FeatureState{{Name: "f", NextTaskID: "t1", NextTaskStatus: "enriched"}},
	}
	rec, err := RecommendNext(snap, "f", nil)
	require.NoError(t, err)
	require.Equal(t, "dev", rec.Action)
	require.Contains(t, rec.Primitive, "--agent dev")
}
