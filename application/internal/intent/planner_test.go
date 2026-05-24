package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
	"github.com/stretchr/testify/require"
)

func TestEvaluateConditions(t *testing.T) {
	fs := FeatureState{HasLocalSpec: false, HasTasks: false, NextTaskStatus: agentflow.StatusPlanned}
	intent := ResolvedIntent{RequiresSync: true, Source: "notion"}
	opts := WorkOptions{}

	require.True(t, EvaluateCondition("source_requires_sync", intent, fs, opts))
	require.True(t, EvaluateCondition("no_local_spec", intent, fs, opts))
	require.True(t, EvaluateCondition("no_tasks", intent, fs, opts))
	require.True(t, EvaluateCondition("task_not_enriched", intent, fs, opts))

	fs.NextTaskStatus = agentflow.StatusImplemented
	require.True(t, EvaluateCondition("implementation_done", intent, fs, opts))

	fs.NextTaskStatus = agentflow.StatusVerifyFailed
	require.True(t, EvaluateCondition("verification_failed", intent, fs, opts))

	opts.NoReview = true
	require.False(t, EvaluateCondition("review_enabled", intent, fs, opts))
}

func TestPlannerDevelopSteps(t *testing.T) {
	p := &DefaultPlanner{}
	cfg := &config.Config{Work: config.WorkConfig{AutoVerify: true, DefaultAgent: "cursor"}}
	snap := StateSnapshot{Features: []FeatureState{
		{Name: "billing-v2", HasLocalSpec: true, HasTasks: true, NextTaskID: "task-003", NextTaskStatus: agentflow.StatusPlanned},
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
