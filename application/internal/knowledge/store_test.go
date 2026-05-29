package knowledge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestGraphStoreUpsertAndLoad(t *testing.T) {
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

	loaded, err := store.LoadGraph(ctx)
	require.NoError(t, err)
	require.Len(t, loaded.Nodes, len(graph.Nodes))
	require.Len(t, loaded.Edges, len(graph.Edges))

	got, err := store.GetNode(ctx, "flow:onboarding")
	require.NoError(t, err)
	require.Equal(t, "onboarding", got.Name)

	_, err = os.Stat(filepath.Join(repo, ".asagiri", "knowledge", "graph.sqlite"))
	require.NoError(t, err)
}

func TestGraphStorePropertiesRoundTrip(t *testing.T) {
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	node := knowledge.GraphNode{
		ID:         "flow:onboarding",
		Type:       knowledge.NodeTypeFlow,
		Name:       "onboarding",
		Properties: map[string]any{"product": "workspace-saas", "steps": 3},
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "store_test"},
		Confidence: 1,
	}
	require.NoError(t, store.UpsertNode(ctx, node))

	got, err := store.GetNode(ctx, node.ID)
	require.NoError(t, err)
	require.Equal(t, "workspace-saas", got.Properties["product"])
	require.EqualValues(t, 3, got.Properties["steps"])

	require.NoError(t, store.SetIndexMetadata(ctx, "build", map[string]any{"nodes": 1}))
	meta, err := store.GetIndexMetadata(ctx, "build")
	require.NoError(t, err)
	require.EqualValues(t, 1, meta["nodes"])
}

func TestGraphStoreGetEdgeRoundTrip(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	from := knowledge.GraphNode{
		ID:         "flow:onboarding",
		Type:       knowledge.NodeTypeFlow,
		Name:       "onboarding",
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "store_test"},
		Confidence: 1,
	}
	to := knowledge.GraphNode{
		ID:         "action:invite_member",
		Type:       knowledge.NodeTypeAction,
		Name:       "invite_member",
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "store_test"},
		Confidence: 1,
	}
	require.NoError(t, store.UpsertNode(ctx, from))
	require.NoError(t, store.UpsertNode(ctx, to))

	edge := knowledge.GraphEdge{
		ID:         knowledge.EdgeID(knowledge.EdgeTypeRequires, from.ID, to.ID),
		From:       from.ID,
		To:         to.ID,
		Type:       knowledge.EdgeTypeRequires,
		Properties: map[string]any{"critical": true},
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "store_test"},
		Confidence: 0.9,
	}
	require.NoError(t, store.UpsertEdge(ctx, edge))

	got, err := store.GetEdge(ctx, edge.ID)
	require.NoError(t, err)
	require.Equal(t, edge.From, got.From)
	require.Equal(t, true, got.Properties["critical"])

	edges, err := store.ListEdges(ctx, knowledge.EdgeFilter{
		Type:       knowledge.EdgeTypeRequires,
		FromNodeID: from.ID,
	})
	require.NoError(t, err)
	require.Len(t, edges, 1)
	require.Equal(t, edge.ID, edges[0].ID)
}

func TestStoreRejectsInvalidIDsAtBoundaries(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	_, err = store.GetNode(ctx, "flow:..segment")
	require.ErrorIs(t, err, knowledge.ErrInvalidNodeID)

	_, err = store.ListNodes(ctx, knowledge.NodeFilter{ID: "flow:..segment"})
	require.ErrorIs(t, err, knowledge.ErrInvalidNodeID)

	_, err = store.ListEdges(ctx, knowledge.EdgeFilter{FromNodeID: "flow:has/slash"})
	require.ErrorIs(t, err, knowledge.ErrInvalidNodeID)

	err = store.SetIndexMetadata(ctx, "../escape", map[string]any{"x": 1})
	require.Error(t, err)
}

func TestOpenStoreRejectsRepoRootWithParentSegments(t *testing.T) {
	_, err := knowledge.OpenStore(filepath.Join("..", "outside"))
	require.Error(t, err)
}

func TestUpsertEdgeRequiresExistingNodes(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	edge := knowledge.GraphEdge{
		ID:         knowledge.EdgeID(knowledge.EdgeTypeRequires, "flow:onboarding", "action:invite_member"),
		From:       "flow:onboarding",
		To:         "action:invite_member",
		Type:       knowledge.EdgeTypeRequires,
		Source:     knowledge.GraphSource{Kind: "fixture"},
		Confidence: 0.9,
	}
	err = store.UpsertEdge(context.Background(), edge)
	require.ErrorIs(t, err, knowledge.ErrNotFound)
}

func TestStubInterfacesReturnNotImplemented(t *testing.T) {
	_, err := (knowledge.StubBuilder{}).Build(t.Context(), knowledge.BuildRequest{})
	require.Error(t, err)
	_, err = (knowledge.StubSnapshotter{}).Snapshot(t.Context(), knowledge.SnapshotRequest{})
	require.ErrorIs(t, err, knowledge.ErrNotImplemented)
	_, err = (knowledge.StubImpactAnalyzer{}).Analyze(t.Context(), knowledge.ImpactRequest{})
	require.ErrorIs(t, err, knowledge.ErrNotImplemented)
	_, err = (knowledge.StubStalenessDetector{}).Check(t.Context(), t.TempDir())
	require.ErrorIs(t, err, knowledge.ErrNotImplemented)
}
