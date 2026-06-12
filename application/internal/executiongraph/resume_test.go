package executiongraph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultRunnerResumeWithoutCheckpointErrors(t *testing.T) {
	repo := t.TempDir()
	graph := ExecutionGraph{
		ID:        "graph-2026-05-29-nocp01",
		Product:   "p",
		Flow:      "f",
		Status:    GraphStatusPaused,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Strategy:  Strategy{MaxParallel: 1, StopOnRisk: RiskLevelHigh},
		Nodes: []GraphNode{
			{ID: "implement-a", Type: NodeTypeImplementation, Status: NodeStatusPending},
		},
	}
	require.NoError(t, graph.Validate())

	repoObj := NewRepository(repo)
	sched := ExecutionSchedule{GraphID: graph.ID, ParallelGroups: [][]string{{"implement-a"}}}
	_, err := repoObj.SaveAll(graph, &sched)
	require.NoError(t, err)

	runner := NewRunner(repo)
	_, err = runner.Resume(context.Background(), graph.ID, RunOptions{})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoCheckpoint)
}
