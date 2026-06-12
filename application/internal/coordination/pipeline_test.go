package coordination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestDefaultPipelineMapGraphReplayOrder(t *testing.T) {
	p := &coordination.DefaultPipeline{Roles: coordination.DefaultPipelineRoles}
	graph := validGraph()
	graph.Nodes = []executiongraph.GraphNode{
		{ID: "validate", Type: executiongraph.NodeTypeValidation},
		{ID: "implement", Type: executiongraph.NodeTypeImplementation},
		{ID: "investigate", Type: executiongraph.NodeTypeInvestigation},
		{ID: "review", Type: executiongraph.NodeTypeReview},
	}

	steps := p.MapGraph(graph)
	require.Len(t, steps, 4)
	require.Equal(t, coordination.RoleInvestigator, steps[0].Role)
	require.Equal(t, coordination.RoleImplementer, steps[1].Role)
	require.Equal(t, []string{"investigate"}, steps[0].NodeIDs)
	require.Equal(t, []string{"implement"}, steps[1].NodeIDs)
	require.Equal(t, "investigate", steps[0].Results[0].NodeID)
}
