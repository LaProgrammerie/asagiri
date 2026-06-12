package executiongraph_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestInferSafeParallelism(t *testing.T) {
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	body, err := os.ReadFile(filepath.Join("..", "knowledge", "testdata", "knowledge-graph", "onboarding-flow", "graph.json"))
	require.NoError(t, err)
	graph, err := knowledge.ParseJSON(body)
	require.NoError(t, err)

	ctx := context.Background()
	for _, node := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, node))
	}
	for _, edge := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, edge))
	}

	execGraph := executiongraph.ExecutionGraph{
		Flow:     "onboarding",
		Strategy: executiongraph.Strategy{MaxParallel: 4},
		Nodes: []executiongraph.GraphNode{
			{ID: "impl-a", Type: executiongraph.NodeTypeImplementation, Outputs: []string{"internal/invitation/invitation_service.go"}},
			{ID: "impl-b", Type: executiongraph.NodeTypeImplementation, Outputs: []string{"internal/invitation/invitation_service.go"}},
		},
	}
	par, reduced, err := executiongraph.InferSafeParallelism(ctx, repo, execGraph)
	require.NoError(t, err)
	require.True(t, reduced)
	require.Equal(t, 1, par)
}

func TestSuggestKnowledgeDeps(t *testing.T) {
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	body, err := os.ReadFile(filepath.Join("..", "knowledge", "testdata", "knowledge-graph", "onboarding-flow", "graph.json"))
	require.NoError(t, err)
	graph, err := knowledge.ParseJSON(body)
	require.NoError(t, err)

	ctx := context.Background()
	for _, node := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, node))
	}
	for _, edge := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, edge))
	}

	deps, err := executiongraph.SuggestKnowledgeDeps(ctx, store, "action:invite_member")
	require.NoError(t, err)
	require.NotEmpty(t, deps)
}
