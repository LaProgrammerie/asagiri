package bus

import (
	"context"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/stretchr/testify/require"
)

func TestHandleGetRunDetailAggregates(t *testing.T) {
	t.Parallel()
	created := time.Date(2026, 5, 31, 9, 0, 0, 0, time.UTC)
	updated := created.Add(30 * time.Minute)

	qb := NewQueryBus(Deps{
		StateOpen: func(string) (stateStore, error) {
			return &mockStateStore{
				runs: []sqlite.Run{{
					ID:        "run-1",
					Feature:   "cockpit-consolidation",
					Status:    "running",
					StepsJSON: `[{"name":"spec","status":"done"},{"name":"verify","status":"running"}]`,
					CreatedAt: created,
					UpdatedAt: updated,
				}},
				tasks: []sqlite.Task{{
					ID: "task-1", RunID: "run-1", WorktreePath: ".asagiri/worktrees/run-1",
				}},
				metric: &telemetry.RunMetric{RunID: "run-1", ActualCostCents: 142},
			}, nil
		},
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{}, nil
		},
	})

	res, err := qb.Query(context.Background(), GetRunDetailQuery{RunID: "run-1"})
	require.NoError(t, err)
	detail, ok := res.(RunDetail)
	require.True(t, ok)

	require.Equal(t, "run-1", detail.ID)
	require.Equal(t, "cockpit-consolidation", detail.Feature)
	require.Equal(t, "running", detail.Status)
	require.Equal(t, ".asagiri/worktrees/run-1", detail.Worktree)
	require.Len(t, detail.Pipeline, 2)
	require.Equal(t, "spec", detail.Pipeline[0].ID)
	require.Equal(t, "done", detail.Pipeline[0].Status)
	require.Equal(t, "verify: running", detail.Validation)
	require.InDelta(t, 1.42, detail.CostEUR, 0.001)
	require.Equal(t, created, detail.CreatedAt)
}

func TestHandleGetRunDetailEmptyID(t *testing.T) {
	t.Parallel()
	qb := NewQueryBus(Deps{})
	res, err := qb.Query(context.Background(), GetRunDetailQuery{RunID: "  "})
	require.NoError(t, err)
	detail, ok := res.(RunDetail)
	require.True(t, ok)
	require.Equal(t, "no run selected", detail.Warning)
}

func TestHandleGetRunDetailNotFound(t *testing.T) {
	t.Parallel()
	qb := NewQueryBus(Deps{
		StateOpen: func(string) (stateStore, error) { return &mockStateStore{}, nil },
		RuntimeOpen: func(string) (runtimeStore, error) { return &mockRuntimeStore{}, nil },
	})
	res, err := qb.Query(context.Background(), GetRunDetailQuery{RunID: "missing"})
	require.NoError(t, err)
	detail, ok := res.(RunDetail)
	require.True(t, ok)
	require.Equal(t, "run not found", detail.Warning)
}

func TestFilterRunAgents(t *testing.T) {
	t.Parallel()
	all := []ActiveAgentSummary{
		{Role: "implementer", AgentRef: "cursor", FlowID: "cockpit-consolidation", Status: "running"},
		{Role: "reviewer", AgentRef: "gpt", FlowID: "other-feature", Status: "idle"},
	}
	matched := filterRunAgents(all, "cockpit-consolidation", "run-1")
	require.Len(t, matched, 1)
	require.Equal(t, "cursor", matched[0].AgentRef)
	require.Empty(t, filterRunAgents(all, "unknown", "run-99"))
}
