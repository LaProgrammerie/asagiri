package trust_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/stretchr/testify/require"
)

func TestAssessTrustFromRepo(t *testing.T) {
	repo := seedGraph(t, "onboarding-flow")
	assess, ok, err := trust.AssessTrustFromRepo(context.Background(), repo, "onboarding", "invite_member")
	require.NoError(t, err)
	require.True(t, ok)
	require.Greater(t, assess.FlowIntegrity, 0.0)
}

func TestBlastRadiusFromGraph(t *testing.T) {
	repo := seedGraph(t, "onboarding-flow")
	report, result, err := trust.BlastRadiusFromGraph(context.Background(), repo, knowledge.ImpactRequest{
		Flow:   "onboarding",
		Action: "invite_member",
	})
	require.NoError(t, err)
	require.NotNil(t, report)
	require.GreaterOrEqual(t, report.FlowsImpacted, 1)
	require.NotEmpty(t, result.ImpactedAPIs)
}

func seedGraph(t *testing.T, scenario string) string {
	t.Helper()
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	path := filepath.Join("..", "knowledge", "testdata", "knowledge-graph", scenario, "graph.json")
	body, err := os.ReadFile(path)
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
	return repo
}
