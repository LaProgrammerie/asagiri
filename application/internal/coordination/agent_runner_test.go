package coordination_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestMarkerNodeAgentRunnerWritesDoneFile(t *testing.T) {
	repo := t.TempDir()
	runner := &coordination.MarkerNodeAgentRunner{}
	graph := executiongraph.ExecutionGraph{ID: "graph-2026-05-29-marker01"}
	node := executiongraph.GraphNode{ID: "node-a", Type: executiongraph.NodeTypeImplementation}
	err := runner.Run(context.Background(), coordination.NodeRunRequest{
		RepoRoot:   repo,
		Graph:      graph,
		Node:       node,
		WorkingDir: repo,
		Assignment: coordination.AgentAssignment{
			NodeID:   node.ID,
			AgentRef: "local",
			Role:     coordination.RoleImplementer,
		},
	})
	require.NoError(t, err)
	marker := filepath.Join(repo, ".asagiri", "agent-runs", graph.ID, node.ID+".done")
	_, err = os.Stat(marker)
	require.NoError(t, err)
}

func TestNodeExecutorWithSharedIsolationSkipsWorktree(t *testing.T) {
	repo := t.TempDir()
	cfg := config.NewTestConfig("proj")
	cfg.Coordination.DefaultIsolation = string(coordination.IsolationShared)
	exec := coordination.NodeExecutor(repo, cfg)
	graph := executiongraph.ExecutionGraph{ID: "graph-2026-05-29-shared01"}
	node := executiongraph.GraphNode{ID: "n1", Type: executiongraph.NodeTypeInvestigation}
	err := exec(context.Background(), graph, node, executiongraph.CoordinationAssignment{
		NodeID:    node.ID,
		AgentRef:  "local",
		Role:      "investigator",
		Isolation: string(coordination.IsolationShared),
	})
	require.NoError(t, err)
	marker := filepath.Join(repo, ".asagiri", "agent-runs", graph.ID, node.ID+".done")
	require.FileExists(t, marker)
}

func TestRunOptionsWiresCoordinatorAndExecutor(t *testing.T) {
	repo := t.TempDir()
	graph := executiongraph.ExecutionGraph{
		ID:        "graph-2026-05-29-full01",
		Product:   "minimal-product",
		Flow:      "workspace-onboarding",
		Status:    executiongraph.GraphStatusReady,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Strategy:  executiongraph.Strategy{MaxParallel: 2, StopOnRisk: executiongraph.RiskLevelHigh},
		Nodes: []executiongraph.GraphNode{
			{ID: "investigate-onboarding", Type: executiongraph.NodeTypeInvestigation, Risk: executiongraph.RiskLevelLow},
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

	marker := filepath.Join(repo, ".asagiri", "agent-runs", graph.ID, "investigate-onboarding.done")
	require.FileExists(t, marker)
}
