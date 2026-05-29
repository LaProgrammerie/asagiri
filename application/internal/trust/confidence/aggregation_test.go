package confidence

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllDimensionsSix(t *testing.T) {
	require.Len(t, AllDimensions, 6)
	require.Equal(t, DimensionArchitecture, AllDimensions[0])
	require.Equal(t, DimensionRegression, AllDimensions[5])
}

func TestStubAggregatorSixUncoveredZones(t *testing.T) {
	t.Parallel()
	var a StubAggregator
	rep, err := a.Aggregate(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, rep.UncoveredZones, 6)
	for i, d := range AllDimensions {
		require.Contains(t, rep.UncoveredZones[i], string(d))
	}
}

func TestReportScoreFor(t *testing.T) {
	rep := Report{
		Architecture:   0.1,
		Implementation: 0.2,
		FlowIntegrity:  0.3,
		Observability:  0.4,
		Security:       0.5,
		Regression:     0.6,
		Overall:        0.35,
	}
	require.Equal(t, 0.1, rep.ScoreFor(DimensionArchitecture))
	require.Equal(t, 0.6, rep.ScoreFor(DimensionRegression))
	require.Zero(t, rep.ScoreFor(Dimension("unknown")))
}

func TestDefaultWeightsEqualSix(t *testing.T) {
	w := DefaultWeights()
	require.Len(t, w, 6)
	var sum float64
	for _, d := range AllDimensions {
		require.Contains(t, w, d)
		sum += w[d]
	}
	require.InDelta(t, 1.0, sum, 1e-9)
}
