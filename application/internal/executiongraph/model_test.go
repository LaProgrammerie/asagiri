package executiongraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecutionGraphValidate(t *testing.T) {
	valid := ExecutionGraph{
		ID:        "graph-2026-05-27-a1b2c3d4",
		Product:   "workspace-saas",
		Status:    GraphStatusPlanned,
		CreatedAt: "2026-05-27T10:00:00Z",
		Strategy:  Strategy{MaxParallel: 2, StopOnRisk: RiskLevelHigh},
		Nodes: []GraphNode{
			{ID: "a", Type: NodeTypeInvestigation, Risk: RiskLevelLow},
			{ID: "b", Type: NodeTypeImplementation, Risk: RiskLevelMedium},
		},
		Edges: []GraphEdge{
			{From: "a", To: "b", Type: EdgeTypeRequires},
		},
		Checkpoints: []Checkpoint{{After: "a"}},
		Rollback:    &RollbackPlan{Strategy: RollbackStrategyWorktreeReset},
	}
	require.NoError(t, valid.Validate())
}

func TestExecutionGraphValidateErrors(t *testing.T) {
	base := ExecutionGraph{
		ID:        "graph-2026-05-27-a1b2c3d4",
		Product:   "workspace-saas",
		Status:    GraphStatusPlanned,
		CreatedAt: "2026-05-27T10:00:00Z",
		Strategy:  Strategy{MaxParallel: 2},
		Nodes: []GraphNode{
			{ID: "a", Type: NodeTypeInvestigation},
		},
	}

	t.Run("duplicate node id", func(t *testing.T) {
		g := base
		g.Nodes = append(g.Nodes, GraphNode{ID: "a", Type: NodeTypeValidation})
		err := g.Validate()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidGraph)
	})

	t.Run("unknown edge endpoint", func(t *testing.T) {
		g := base
		g.Edges = []GraphEdge{{From: "a", To: "missing", Type: EdgeTypeRequires}}
		err := g.Validate()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidGraph)
	})

	t.Run("invalid max parallel", func(t *testing.T) {
		g := base
		g.Strategy.MaxParallel = 0
		err := g.Validate()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidGraph)
	})

	t.Run("checkpoint unknown node", func(t *testing.T) {
		g := base
		g.Checkpoints = []Checkpoint{{After: "missing"}}
		err := g.Validate()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidGraph)
	})

	t.Run("invalid graph id", func(t *testing.T) {
		g := base
		g.ID = "../escape"
		err := g.Validate()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidGraph)
	})
}

func TestSortedNodesAndEdges(t *testing.T) {
	g := ExecutionGraph{
		Nodes: []GraphNode{
			{ID: "z-node", Type: NodeTypeValidation},
			{ID: "a-node", Type: NodeTypeInvestigation},
		},
		Edges: []GraphEdge{
			{From: "b", To: "c", Type: EdgeTypeValidates},
			{From: "a", To: "b", Type: EdgeTypeRequires},
			{From: "a", To: "b", Type: EdgeTypeBlocks},
		},
	}
	nodes := g.SortedNodes()
	require.Equal(t, "a-node", nodes[0].ID)
	require.Equal(t, "z-node", nodes[1].ID)

	edges := g.SortedEdges()
	require.Equal(t, EdgeTypeBlocks, edges[0].Type)
	require.Equal(t, EdgeTypeRequires, edges[1].Type)
}

func TestParseYAMLFromFixture(t *testing.T) {
	for _, scenario := range goldenScenarios() {
		t.Run(scenario, func(t *testing.T) {
			body := readGoldenFixture(t, scenario)
			graph, err := ParseYAML(body)
			require.NoError(t, err)
			require.NoError(t, graph.Validate())
		})
	}
}
