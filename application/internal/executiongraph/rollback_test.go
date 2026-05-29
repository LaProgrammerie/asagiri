package executiongraph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRollbackStrategyForImplementation(t *testing.T) {
	strategy, ok := rollbackStrategyFor(GraphNode{
		ID:   "implement-workspace-create",
		Type: NodeTypeImplementation,
		Risk: RiskLevelHigh,
	}, nil)
	require.True(t, ok)
	require.Equal(t, RollbackStrategyWorktreeReset, strategy)
}

func TestRollbackStrategyForMigration(t *testing.T) {
	strategy, ok := rollbackStrategyFor(GraphNode{
		ID:   "implement-db-migration",
		Type: NodeTypeImplementation,
		Risk: RiskLevelHigh,
	}, &TaskBinding{Action: "run_migration"})
	require.True(t, ok)
	require.Equal(t, RollbackStrategyMigrationDown, strategy)
}

func TestRollbackStrategyForPublicContract(t *testing.T) {
	strategy, ok := rollbackStrategyFor(GraphNode{
		ID:   "implement-api",
		Type: NodeTypeImplementation,
		Risk: RiskLevelHigh,
	}, &TaskBinding{ContractRef: "POST /api/workspaces"})
	require.True(t, ok)
	require.Equal(t, RollbackStrategyPatchRevert, strategy)
}

func TestPlanRollbackDetectsMissingStrategy(t *testing.T) {
	graph := ExecutionGraph{
		Nodes: []GraphNode{
			{ID: "unknown-risky", Type: NodeTypeEnrichment, Risk: RiskLevelHigh},
		},
	}
	assessment := PlanRollback(graph, nil)
	require.Contains(t, assessment.MissingStrategy, "unknown-risky")
	require.Equal(t, RollbackStrategyManual, assessment.GraphDefault.Strategy)
	require.NotEmpty(t, assessment.Warnings)
}

func TestApplyRollbackEnrichmentSetsNodeAndGraphDefaults(t *testing.T) {
	graph := ExecutionGraph{
		Nodes: []GraphNode{{
			ID:   "implement-invite-member",
			Type: NodeTypeImplementation,
			Risk: RiskLevelHigh,
		}},
	}
	assessment := ApplyRollbackEnrichment(&graph, nil)
	require.Equal(t, RollbackStrategyWorktreeReset, graph.Nodes[0].RollbackStrategy)
	require.Equal(t, RollbackStrategyWorktreeReset, graph.Rollback.Strategy)
	require.Empty(t, assessment.MissingStrategy)
}

func TestGenerateCheckpointsFromGraph(t *testing.T) {
	graph := ExecutionGraph{
		Nodes: []GraphNode{
			{ID: "investigate-onboarding", Type: NodeTypeInvestigation, Risk: RiskLevelLow},
			{ID: "implement-workspace-create", Type: NodeTypeImplementation, Risk: RiskLevelMedium},
			{ID: "verify-onboarding-flow", Type: NodeTypeValidation, Risk: RiskLevelMedium},
		},
		Edges: []GraphEdge{
			{From: "investigate-onboarding", To: "implement-workspace-create", Type: EdgeTypeProducesContextFor},
			{From: "implement-workspace-create", To: "verify-onboarding-flow", Type: EdgeTypeValidates},
		},
	}
	checkpoints := GenerateCheckpoints(graph)
	require.Len(t, checkpoints, 3)
	require.Equal(t, "investigate-onboarding", checkpoints[0].After)
	require.Equal(t, "implement-workspace-create", checkpoints[1].After)
	require.Equal(t, "verify-onboarding-flow", checkpoints[2].After)
}

func TestEnrichGraphIntegration(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	planner := Planner{
		RepoRoot: repo,
		Inferer:  DefaultDependencyInferer{},
		Now: func() time.Time {
			return time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
		},
	}

	graph, err := planner.Build(t.Context(), GraphPlanRequest{
		Product:        "minimal-product",
		Flow:           "workspace-onboarding",
		IncludeReviews: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, graph.Checkpoints)
	require.NotNil(t, graph.Rollback)
	require.Contains(t, []RollbackStrategy{
		RollbackStrategyWorktreeReset,
		RollbackStrategyPatchRevert,
		RollbackStrategyManual,
	}, graph.Rollback.Strategy)

	for _, n := range graph.Nodes {
		require.NotEmpty(t, n.Agent)
		if n.Type == NodeTypeImplementation && n.Risk == RiskLevelHigh {
			require.NotEmpty(t, n.RollbackStrategy)
		}
	}
}
