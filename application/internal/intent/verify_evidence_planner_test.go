package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func verifyEvidenceActiveCfg() *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			DefaultReviewer: "reviewer",
			Gates: config.WorkGatesConfig{
				VerifyEvidence: config.WorkVerifyEvidenceGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
}

func TestRecommendNextVerifiedWithoutVerifyEvidenceHistoryBlocksReview(t *testing.T) {
	cfg := verifyEvidenceActiveCfg()
	snap := StateSnapshot{
		ActiveFeature: "feat",
		Features: []FeatureState{{
			Name: "feat", NextTaskID: "t1", NextTaskStatus: asagiri.StatusVerified,
			VerifyEvidenceGateBlocksReview: true,
		}},
	}
	rec, err := RecommendNext(snap, "feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "verify", rec.Action)
	require.Contains(t, rec.Primitive, "asa verify feat --task t1 --force")
}

func TestRecommendNextVerifiedWithVerifyEvidenceSatisfiedRecommendsReview(t *testing.T) {
	cfg := verifyEvidenceActiveCfg()
	snap := StateSnapshot{
		ActiveFeature: "feat",
		Features: []FeatureState{{
			Name: "feat", NextTaskID: "t1", NextTaskStatus: asagiri.StatusVerified,
			VerifyEvidenceGateBlocksReview: false,
		}},
	}
	rec, err := RecommendNext(snap, "feat", cfg)
	require.NoError(t, err)
	require.Equal(t, "review", rec.Action)
	require.Contains(t, rec.Primitive, "asa review feat --task t1")
}

func TestEvaluateConditionReviewEnabledBlocksWhenVerifyEvidenceUnsatisfied(t *testing.T) {
	fs := FeatureState{VerifyEvidenceGateBlocksReview: true}
	require.False(t, EvaluateCondition("review_enabled", ResolvedIntent{}, fs, WorkOptions{}))

	fs.VerifyEvidenceGateBlocksReview = false
	require.True(t, EvaluateCondition("review_enabled", ResolvedIntent{}, fs, WorkOptions{}))
}

func TestPlannerSkipsReviewWhenVerifyEvidenceBlocks(t *testing.T) {
	p := &DefaultPlanner{}
	cfg := &config.Config{Work: config.WorkConfig{
		AutoVerify:   true,
		AutoReview:   true,
		DefaultAgent: "cursor",
	}}
	snap := StateSnapshot{Features: []FeatureState{{
		Name: "feat", HasLocalSpec: true, HasTasks: true,
		NextTaskID: "t1", NextTaskStatus: asagiri.StatusVerified,
		VerifyEvidenceGateBlocksReview: true,
	}}}
	plan, err := p.BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "feat", TaskID: "t1",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)

	fs := featureState(snap, "feat")
	runnableReview := 0
	for _, step := range plan.Steps {
		if step.Command == "review" && EvaluateCondition(step.Condition, plan.Intent, fs, WorkOptions{}) {
			runnableReview++
		}
	}
	require.Equal(t, 0, runnableReview)
}

func TestPlannerAllowsReviewWhenVerifyEvidenceSatisfied(t *testing.T) {
	p := &DefaultPlanner{}
	cfg := &config.Config{Work: config.WorkConfig{
		AutoVerify:   true,
		AutoReview:   true,
		DefaultAgent: "cursor",
	}}
	snap := StateSnapshot{Features: []FeatureState{{
		Name: "feat", HasLocalSpec: true, HasTasks: true,
		NextTaskID: "t1", NextTaskStatus: asagiri.StatusVerified,
		VerifyEvidenceGateBlocksReview: false,
	}}}
	plan, err := p.BuildPlan(context.Background(), ResolvedIntent{
		Action: IntentDevelop, Feature: "feat", TaskID: "t1",
	}, snap, cfg, WorkOptions{})
	require.NoError(t, err)

	fs := featureState(snap, "feat")
	runnableReview := 0
	for _, step := range plan.Steps {
		if step.Command == "review" && EvaluateCondition(step.Condition, plan.Intent, fs, WorkOptions{}) {
			runnableReview++
		}
	}
	require.Equal(t, 1, runnableReview)
}
