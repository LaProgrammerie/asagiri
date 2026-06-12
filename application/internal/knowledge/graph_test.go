package knowledge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func TestFixtureGraphsValidate(t *testing.T) {
	for _, scenario := range []string{"minimal", "onboarding-flow", "api-events"} {
		t.Run(scenario, func(t *testing.T) {
			graph := loadFixtureGraph(t, scenario)
			require.NoError(t, graph.Validate())
			sorted := graph.SortedNodes()
			require.Equal(t, graph.Nodes[0].ID, sorted[0].ID)
			if len(sorted) > 1 {
				require.True(t, sorted[0].ID < sorted[1].ID)
			}
		})
	}
}

func TestPruneOrphanEdges(t *testing.T) {
	graph := knowledge.KnowledgeGraph{
		Nodes: []knowledge.GraphNode{
			{
				ID:         "symbol:config_ExamplePath",
				Type:       knowledge.NodeTypeSymbol,
				Name:       "ExamplePath",
				Source:     knowledge.GraphSource{Kind: "code"},
				Confidence: 0.9,
			},
			{
				ID:         "test:ExamplePathTest",
				Type:       knowledge.NodeTypeTest,
				Name:       "TestExamplePath",
				Source:     knowledge.GraphSource{Kind: "test"},
				Confidence: 0.9,
			},
		},
		Edges: []knowledge.GraphEdge{
			{
				ID:         knowledge.EdgeID(knowledge.EdgeTypeTests, "symbol:config_ExamplePath", "test:ExamplePathTest"),
				From:       "symbol:config_ExamplePath",
				To:         "test:ExamplePathTest",
				Type:       knowledge.EdgeTypeTests,
				Source:     knowledge.GraphSource{Kind: "test"},
				Confidence: 0.9,
			},
			{
				ID:         knowledge.EdgeID(knowledge.EdgeTypeTests, "symbol:missing", "test:orphan"),
				From:       "symbol:missing",
				To:         "test:orphan",
				Type:       knowledge.EdgeTypeTests,
				Source:     knowledge.GraphSource{Kind: "test"},
				Confidence: 0.9,
			},
		},
	}
	pruned, dropped := knowledge.PruneOrphanEdges(graph)
	require.Len(t, dropped, 1)
	require.Len(t, pruned.Edges, 1)
	require.NoError(t, pruned.Validate())
}

func TestGraphValidateRejectsUnknownEdgeEndpoint(t *testing.T) {
	graph := knowledge.KnowledgeGraph{
		Nodes: []knowledge.GraphNode{{
			ID:         "flow:onboarding",
			Type:       knowledge.NodeTypeFlow,
			Source:     knowledge.GraphSource{Kind: "fixture"},
			Confidence: 1,
		}},
		Edges: []knowledge.GraphEdge{{
			ID:         knowledge.EdgeID(knowledge.EdgeTypeRequires, "flow:onboarding", "action:missing"),
			From:       "flow:onboarding",
			To:         "action:missing",
			Type:       knowledge.EdgeTypeRequires,
			Source:     knowledge.GraphSource{Kind: "fixture"},
			Confidence: 0.9,
		}},
	}
	require.ErrorIs(t, graph.Validate(), knowledge.ErrInvalidGraph)
}

func loadFixtureGraph(t *testing.T, scenario string) knowledge.KnowledgeGraph {
	t.Helper()
	path := filepath.Join("testdata", "knowledge-graph", scenario, "graph.json")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	graph, err := knowledge.ParseJSON(body)
	require.NoError(t, err)
	return graph
}
