package executiongraph

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAssessRollbackImpactAndExecute(t *testing.T) {
	repoRoot := t.TempDir()
	graph := ExecutionGraph{
		ID:        "graph-2026-05-29-abcdef01",
		Product:   "workspace-saas",
		Flow:      "onboarding",
		Status:    GraphStatusRunning,
		CreatedAt: "2026-05-29T10:00:00Z",
		Strategy:  Strategy{MaxParallel: 1},
		Nodes: []GraphNode{
			{
				ID:               "implement-invite",
				Type:             NodeTypeImplementation,
				Risk:             RiskLevelHigh,
				Status:           NodeStatusRunning,
				RollbackStrategy: RollbackStrategyWorktreeReset,
			},
		},
		Rollback: &RollbackPlan{Strategy: RollbackStrategyWorktreeReset, PreserveReports: true},
	}
	repo := NewRepository(repoRoot)
	_, _, err := repo.Save(graph)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(repo.graphDir(graph.ID), "plan.md"), []byte("# plan"), 0o644))

	impact, err := AssessRollbackImpact(repoRoot, graph.ID)
	require.NoError(t, err)
	require.Contains(t, impact.Title, graph.ID)
	require.NotEmpty(t, impact.ImpactLines)
	require.True(t, impact.RollbackPossible)

	result, err := ExecuteGraphRollback(context.Background(), repoRoot, graph.ID, false)
	require.NoError(t, err)
	require.Equal(t, GraphStatusRolledBack, result.Status)
	require.Equal(t, 1, result.NodesRolledBack)

	loaded, err := repo.Load(graph.ID)
	require.NoError(t, err)
	require.Equal(t, GraphStatusRolledBack, loaded.Status)
	require.Equal(t, NodeStatusRolledBack, loaded.Nodes[0].Status)
}

func TestExecuteGraphRollbackDryRunDoesNotPersist(t *testing.T) {
	repoRoot := t.TempDir()
	graph := ExecutionGraph{
		ID:        "graph-2026-05-29-abcdef02",
		Product:   "workspace-saas",
		Status:    GraphStatusRunning,
		CreatedAt: "2026-05-29T10:00:00Z",
		Strategy:  Strategy{MaxParallel: 1},
		Nodes: []GraphNode{
			{ID: "n1", Type: NodeTypeImplementation, Status: NodeStatusRunning, Risk: RiskLevelLow},
		},
	}
	repo := NewRepository(repoRoot)
	_, _, err := repo.Save(graph)
	require.NoError(t, err)

	_, err = ExecuteGraphRollback(context.Background(), repoRoot, graph.ID, true)
	require.NoError(t, err)

	body, err := os.ReadFile(filepath.Join(repo.graphDir(graph.ID), "execution-graph.yaml"))
	require.NoError(t, err)
	var persisted ExecutionGraph
	require.NoError(t, yaml.Unmarshal(body, &persisted))
	require.Equal(t, GraphStatusRunning, persisted.Status)
}
