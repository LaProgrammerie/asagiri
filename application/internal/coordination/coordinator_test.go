package coordination_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestDefaultCoordinatorAssignsNodes(t *testing.T) {
	graph := validGraph()
	graph.Nodes = []executiongraph.GraphNode{
		{ID: "investigate", Type: executiongraph.NodeTypeInvestigation},
		{ID: "implement", Type: executiongraph.NodeTypeImplementation, Risk: executiongraph.RiskLevelMedium},
	}

	coord := &coordination.DefaultCoordinator{
		Assigner: &coordination.DefaultAssigner{
			Config: coordination.AssignerConfig{
				DefaultIsolation: coordination.IsolationIsolatedWorktree,
				Assignment: map[string]string{
					"investigation":         "local",
					"implementation.medium": "cursor",
				},
			},
		},
		Policies: coordination.PolicyEvaluator{
			Policies: coordination.CoordinationPolicies{MaxParallelAgents: 2},
		},
	}

	require.NoError(t, coord.Coordinate(context.Background(), graph))
	result := coord.LastResult()
	require.Len(t, result.Assignments, 2)
	require.Equal(t, "local", result.Graph.Nodes[0].Agent)
	require.Equal(t, "cursor", result.Graph.Nodes[1].Agent)
}

func TestDefaultCoordinatorPolicyViolation(t *testing.T) {
	g := validGraph()
	g.Strategy.MaxParallel = 10
	coord := &coordination.DefaultCoordinator{
		Assigner: &coordination.DefaultAssigner{},
		Policies: coordination.PolicyEvaluator{
			Policies: coordination.CoordinationPolicies{MaxParallelAgents: 2},
		},
	}
	err := coord.Coordinate(context.Background(), g)
	require.Error(t, err)
	require.ErrorIs(t, err, coordination.ErrPolicyViolation)
}
