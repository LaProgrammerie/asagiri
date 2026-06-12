package coordination_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestRunnerCoordinatorWiresGraphExecution(t *testing.T) {
	repo := t.TempDir()
	graph := executiongraph.ExecutionGraph{
		ID:        "graph-2026-05-29-coord01",
		Product:   "minimal-product",
		Flow:      "workspace-onboarding",
		Status:    executiongraph.GraphStatusReady,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Strategy:  executiongraph.Strategy{MaxParallel: 2, StopOnRisk: executiongraph.RiskLevelHigh},
		Nodes: []executiongraph.GraphNode{
			{ID: "investigate-onboarding", Type: executiongraph.NodeTypeInvestigation, Risk: executiongraph.RiskLevelLow},
			{ID: "implement-click-get-started", Type: executiongraph.NodeTypeImplementation, Risk: executiongraph.RiskLevelMedium},
		},
		Edges: []executiongraph.GraphEdge{
			{From: "investigate-onboarding", To: "implement-click-get-started", Type: executiongraph.EdgeTypeProducesContextFor},
		},
	}
	require.NoError(t, graph.Validate())

	sched, err := executiongraph.DefaultScheduler{}.Schedule(t.Context(), executiongraph.ScheduleRequest{Graph: graph})
	require.NoError(t, err)

	cfg := config.NewTestConfig("proj")
	cfg.Coordination.DefaultIsolation = string(coordination.IsolationShared)
	runner := executiongraph.NewRunner(repo)
	result, err := runner.Run(t.Context(), graph, sched, coordination.RunOptions(cfg, repo, nil))
	require.NoError(t, err)
	require.Equal(t, executiongraph.GraphStatusCompleted, result.Status)

	loaded, err := executiongraph.NewRepository(repo).Load(graph.ID)
	require.NoError(t, err)
	require.NotEmpty(t, loaded.Nodes[0].Agent)
}
