package coordination_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestApplyGraphAgentRouting(t *testing.T) {
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

	asg := coordination.AgentAssignment{Role: coordination.RoleImplementer}
	updated, err := coordination.ApplyGraphAgentRouting(ctx, store, "onboarding", "invite_member",
		executiongraph.GraphNode{Risk: executiongraph.RiskLevelHigh}, asg)
	require.NoError(t, err)
	require.NotEqual(t, coordination.RoleImplementer, updated.Role)
}

func TestEnrichContextFromGraph(t *testing.T) {
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

	pack := coordination.ContextPack{Flow: "onboarding"}
	enriched, err := coordination.EnrichContextFromGraph(ctx, pack, store, "onboarding", "invite_member")
	require.NoError(t, err)
	require.Contains(t, enriched.Files, "internal/invitation/invitation_service.go")
}
