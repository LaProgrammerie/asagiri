package coordination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestReduceContextFiltersFilesAndInjectsMetadata(t *testing.T) {
	graph := validGraph()
	graph.Product = "saas"
	graph.Flow = "onboarding"
	graph.Nodes = append(graph.Nodes, executiongraph.GraphNode{
		ID:      "inv",
		Type:    executiongraph.NodeTypeInvestigation,
		Outputs: []string{"docs/spec.md"},
	})

	node := executiongraph.GraphNode{
		ID:   "impl",
		Type: executiongraph.NodeTypeImplementation,
		Outputs: []string{
			"src/a.go", "../escape.go", "src/b.go", "src/c.go", "src/d.go",
		},
	}

	pack, before, after := coordination.ReduceContext(node, graph, coordination.HandoffHints{MaxFiles: 2})
	require.Equal(t, "saas", pack.Product)
	require.Equal(t, "onboarding", pack.Flow)
	require.Len(t, pack.Files, 2)
	require.NotContains(t, pack.Files, "../escape.go")
	require.Contains(t, pack.InvestigationOutputs, "docs/spec.md")
	require.GreaterOrEqual(t, before, after)
	require.Less(t, after, before+len(node.Outputs)*8)
}
