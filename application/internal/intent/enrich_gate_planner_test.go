package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestRecommendNextEnrichedWithoutEnrichHistoryBlocksDev(t *testing.T) {
	cfg := &config.Config{
		Work: config.WorkConfig{
			DefaultEnricher: "local-rag",
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{Enabled: true, Mode: config.GovernanceModePerTask},
			},
		},
	}
	snap := StateSnapshot{
		ActiveFeature: "feat",
		Features: []FeatureState{{
			Name: "feat", NextTaskID: "t1", NextTaskStatus: asagiri.StatusEnriched,
			EnrichGateBlocksDev: true,
		}},
	}
	rec, err := RecommendNext(snap, "feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "enrich", rec.Action)
}

func TestRecommendNextEnrichGateBlocksDev(t *testing.T) {
	cfg := &config.Config{
		Work: config.WorkConfig{
			DefaultEnricher: "local-rag",
			DefaultAgent:    "cursor",
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{Enabled: true, Mode: config.GovernanceModePerTask},
			},
		},
	}
	snap := StateSnapshot{
		ActiveFeature: "feat",
		Features: []FeatureState{{
			Name: "feat", NextTaskID: "t1", NextTaskStatus: asagiri.StatusPlanned,
			EnrichGateBlocksDev: true,
		}},
	}
	rec, err := RecommendNext(snap, "feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "enrich", rec.Action)
	require.Contains(t, rec.Primitive, "asa enrich feat --task t1")
	require.Contains(t, rec.Primitive, "--force")
	require.Contains(t, rec.Primitive, "--agent local-rag")
}

func TestRecommendNextPlannedEnrichGateSatisfiedRecommendsDev(t *testing.T) {
	cfg := &config.Config{Work: config.WorkConfig{DefaultAgent: "cursor"}}
	snap := StateSnapshot{
		ActiveFeature: "feat",
		Features: []FeatureState{{
			Name: "feat", NextTaskID: "t1", NextTaskStatus: asagiri.StatusPlanned,
			EnrichGateBlocksDev: false,
		}},
	}
	rec, err := RecommendNext(snap, "feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "dev", rec.Action)
}

func TestEvaluateConditionTaskReadyForDev(t *testing.T) {
	fs := FeatureState{NextTaskStatus: asagiri.StatusPlanned, EnrichGateBlocksDev: true}
	require.False(t, EvaluateCondition("task_ready_for_dev", ResolvedIntent{}, fs, WorkOptions{}))

	fs.EnrichGateBlocksDev = false
	require.True(t, EvaluateCondition("task_ready_for_dev", ResolvedIntent{}, fs, WorkOptions{}))

	fs.NextTaskStatus = asagiri.StatusEnriched
	fs.EnrichGateBlocksDev = true
	require.False(t, EvaluateCondition("task_ready_for_dev", ResolvedIntent{}, fs, WorkOptions{}))
}

func TestPlannerSkipsDevWhenEnrichGateBlocks(t *testing.T) {
	p := &DefaultPlanner{}
	cfg := &config.Config{Work: config.WorkConfig{AutoVerify: false, DefaultAgent: "cursor"}}
	snap := StateSnapshot{Features: []FeatureState{{
		Name: "feat", HasLocalSpec: true, HasTasks: true,
		NextTaskID: "t1", NextTaskStatus: asagiri.StatusPlanned,
		EnrichGateBlocksDev: true,
	}}}
	plan, err := p.BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "feat", TaskID: "t1",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)

	fs := featureState(snap, "feat")
	runnableDev := 0
	runnableEnrich := 0
	for _, step := range plan.Steps {
		if !EvaluateCondition(step.Condition, plan.Intent, fs, WorkOptions{}) {
			continue
		}
		switch step.Command {
		case "dev":
			runnableDev++
		case "enrich":
			runnableEnrich++
		}
	}
	require.Equal(t, 0, runnableDev)
	require.Equal(t, 1, runnableEnrich)
}

func TestPlannerAllowsDevWhenEnrichGateSatisfied(t *testing.T) {
	p := &DefaultPlanner{}
	cfg := &config.Config{Work: config.WorkConfig{AutoVerify: false, DefaultAgent: "cursor"}}
	snap := StateSnapshot{Features: []FeatureState{{
		Name: "feat", HasLocalSpec: true, HasTasks: true,
		NextTaskID: "t1", NextTaskStatus: asagiri.StatusEnriched,
		EnrichGateBlocksDev: false,
	}}}
	plan, err := p.BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "feat", TaskID: "t1",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)

	fs := featureState(snap, "feat")
	runnableDev := 0
	for _, step := range plan.Steps {
		if step.Command == "dev" && EvaluateCondition(step.Condition, plan.Intent, fs, WorkOptions{}) {
			runnableDev++
		}
	}
	require.Equal(t, 1, runnableDev)
}
