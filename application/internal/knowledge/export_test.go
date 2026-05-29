package knowledge_test

import (
	"context"
	"os"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestExportRepoGraphWritesParseableJSON(t *testing.T) {
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	graph := loadFixtureGraph(t, "minimal")
	ctx := context.Background()
	for _, node := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, node))
	}
	for _, edge := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, edge))
	}

	require.NoError(t, knowledge.ExportRepoGraph(repo))

	body, err := os.ReadFile(knowledge.JSONPath(repo))
	require.NoError(t, err)

	exported, err := knowledge.ParseJSON(body)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())
	require.Len(t, exported.Nodes, len(graph.Nodes))
	require.Len(t, exported.Edges, len(graph.Edges))
}

func TestPersistGraphJSONRejectsInvalidGraph(t *testing.T) {
	err := knowledge.PersistGraphJSON(t.TempDir(), knowledge.KnowledgeGraph{
		Edges: []knowledge.GraphEdge{{
			ID:         "requires:flow_onboarding>action_invite_member",
			From:       "flow:onboarding",
			To:         "action:invite_member",
			Type:       knowledge.EdgeTypeRequires,
			Source:     knowledge.GraphSource{Kind: "fixture"},
			Confidence: 0.9,
		}},
	})
	require.ErrorIs(t, err, knowledge.ErrInvalidGraph)
}
