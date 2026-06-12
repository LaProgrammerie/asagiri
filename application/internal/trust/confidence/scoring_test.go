package confidence

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultScorerDimensions(t *testing.T) {
	raw, err := DefaultScorer{}.Score(context.Background(), []DetailedCheck{
		{Type: "static-analysis", Confidence: 0.9},
		{Type: "contracts", Confidence: 0.8, Findings: []FindingInput{{Severity: "warning", Category: "architecture.contract"}}},
		{Type: "flows", Confidence: 0.85},
		{Type: "tests", Confidence: 0.7},
	})
	require.NoError(t, err)
	require.Contains(t, raw, DimensionArchitecture)
	require.Contains(t, raw, DimensionImplementation)
	require.Greater(t, raw[DimensionArchitecture], 0.0)
}

func TestRealAggregatorWithChecks(t *testing.T) {
	rep, err := NewRealAggregator().AggregateDetailed(context.Background(), []DetailedCheck{
		{Type: "static-analysis", Confidence: 0.9},
		{Type: "contracts", Confidence: 0.75},
		{Type: "flows", Confidence: 0.85},
		{Type: "tests", Confidence: 0.6},
	})
	require.NoError(t, err)
	require.Greater(t, rep.Overall, 0.0)
	require.NotContains(t, rep.Limits[0], "skeleton")
}

func TestRealAggregatorPopulatesSixDimensions(t *testing.T) {
	rep, err := NewRealAggregator().AggregateDetailed(context.Background(), []DetailedCheck{
		{Type: "static-analysis", Confidence: 0.9},
		{Type: "contracts", Confidence: 0.75, Findings: []FindingInput{{Severity: "warning", Category: "architecture.contract"}}},
		{Type: "flows", Confidence: 0.85},
		{Type: "tests", Confidence: 0.6},
	})
	require.NoError(t, err)
	for _, d := range AllDimensions {
		require.Greater(t, rep.ScoreFor(d), 0.0, "dimension %s", d)
	}
	require.Greater(t, rep.Overall, 0.0)
	require.Empty(t, rep.UncoveredZones)
	require.ElementsMatch(t, []string{"observability", "security"}, rep.InferredDimensions)
	require.LessOrEqual(t, rep.Observability, InferredDimensionCap)
	require.LessOrEqual(t, rep.Security, InferredDimensionCap)
}

func TestRealAggregatorDedicatedObservabilitySecurityUncapped(t *testing.T) {
	rep, err := NewRealAggregator().AggregateDetailed(context.Background(), []DetailedCheck{
		{Type: "static-analysis", Confidence: 0.9},
		{Type: "contracts", Confidence: 0.75},
		{Type: "flows", Confidence: 0.85},
		{Type: "observability", Confidence: 0.8},
		{Type: "security", Confidence: 0.82},
		{Type: "tests", Confidence: 0.6},
	})
	require.NoError(t, err)
	require.NotContains(t, rep.InferredDimensions, "observability")
	require.NotContains(t, rep.InferredDimensions, "security")
	require.Greater(t, rep.Observability, InferredDimensionCap)
	require.Greater(t, rep.Security, InferredDimensionCap)
}

func TestDefaultScorerSkipsSkippedChecks(t *testing.T) {
	withSkip := []DetailedCheck{
		{Type: "static-analysis", Status: checkStatusSkipped, Confidence: 0},
		{Type: "contracts", Confidence: 0.8},
		{Type: "flows", Confidence: 0.85},
		{Type: "tests", Confidence: 0.7},
	}
	withoutSkip := []DetailedCheck{
		{Type: "static-analysis", Confidence: 0.9},
		{Type: "contracts", Confidence: 0.8},
		{Type: "flows", Confidence: 0.85},
		{Type: "tests", Confidence: 0.7},
	}
	skippedRaw, err := DefaultScorer{}.Score(context.Background(), withSkip)
	require.NoError(t, err)
	fullRaw, err := DefaultScorer{}.Score(context.Background(), withoutSkip)
	require.NoError(t, err)
	require.Greater(t, fullRaw[DimensionImplementation], skippedRaw[DimensionImplementation])
}

func TestDefaultScorerCoversSixDimensions(t *testing.T) {
	raw, err := DefaultScorer{}.Score(context.Background(), []DetailedCheck{
		{Type: "static-analysis", Confidence: 0.9},
		{Type: "contracts", Confidence: 0.8},
		{Type: "flows", Confidence: 0.85},
		{Type: "tests", Confidence: 0.7},
	})
	require.NoError(t, err)
	require.Len(t, raw, 6)
	for _, d := range AllDimensions {
		require.Contains(t, raw, d)
		require.Greater(t, raw[d], 0.0)
	}
}
