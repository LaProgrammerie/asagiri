package coordination_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestMemoryBudgetTrackerAggregatesByRole(t *testing.T) {
	tracker := coordination.NewMemoryBudgetTracker()
	ctx := context.Background()
	graphID := "graph-2026-05-29-budget01"

	require.NoError(t, tracker.Record(ctx, graphID, coordination.AgentAssignment{
		Role: coordination.RoleImplementer,
	}, 0.14, 1000, 500))
	require.NoError(t, tracker.Record(ctx, graphID, coordination.AgentAssignment{
		Role: coordination.RoleReviewer,
	}, 0.05, 200, 100))

	summary, err := tracker.Summary(ctx, graphID)
	require.NoError(t, err)
	require.InDelta(t, 0.19, summary.TotalEUR, 0.001)
	require.Len(t, summary.Agents, 2)
}
