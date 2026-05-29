package knowledge_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestQuerierBFSRespectsDepthAndLimit(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	graph := loadFixtureGraph(t, "onboarding-flow")
	ctx := context.Background()
	for _, node := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, node))
	}
	for _, edge := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, edge))
	}

	q := knowledge.NewQuerier(store)
	result, err := q.Query(ctx, knowledge.GraphQuery{
		StartID:  "flow:onboarding",
		MaxDepth: 3,
		Limit:    4,
	})
	require.NoError(t, err)
	require.Len(t, result.Nodes, 4)
	require.NotEmpty(t, result.Edges)

	shallow, _, err := q.BFS(ctx, "flow:onboarding", knowledge.BFSOptions{MaxDepth: 1, Limit: 10})
	require.NoError(t, err)
	require.Len(t, shallow, 2)
}

func TestQuerierRejectsInvalidNodeID(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	q := knowledge.NewQuerier(store)
	_, err = q.Query(context.Background(), knowledge.GraphQuery{NodeID: "flow:..segment"})
	require.ErrorIs(t, err, knowledge.ErrInvalidNodeID)

	_, err = q.Query(context.Background(), knowledge.GraphQuery{StartID: "flow:has/slash"})
	require.ErrorIs(t, err, knowledge.ErrInvalidNodeID)
}

func TestQuerierFilterByTypeAndPath(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	graph := loadFixtureGraph(t, "onboarding-flow")
	ctx := context.Background()
	for _, node := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, node))
	}
	for _, edge := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, edge))
	}

	q := knowledge.NewQuerier(store)
	result, err := q.Query(ctx, knowledge.GraphQuery{
		NodeType: knowledge.NodeTypeTest,
	})
	require.NoError(t, err)
	require.Len(t, result.Nodes, 1)
	require.Equal(t, "test:InvitationServiceTest", result.Nodes[0].ID)

	result, err = q.Query(ctx, knowledge.GraphQuery{
		PathPrefix: "internal/invitation/",
		NodeType:   knowledge.NodeTypeTest,
	})
	require.NoError(t, err)
	require.Len(t, result.Nodes, 1)

	result, err = q.Query(ctx, knowledge.GraphQuery{
		NodeID: "action:invite_member",
	})
	require.NoError(t, err)
	require.Len(t, result.Nodes, 1)
	require.Equal(t, "action:invite_member", result.Nodes[0].ID)

	result, err = q.Query(ctx, knowledge.GraphQuery{
		FromNodeID: "flow:onboarding",
		EdgeType:   knowledge.EdgeTypeRequires,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Edges)
	require.Equal(t, "flow:onboarding", result.Edges[0].From)
}
