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

func TestAssessTrustFromGraphOnboarding(t *testing.T) {
	repo := loadOnboardingStore(t)
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	assess, ok, err := knowledge.AssessTrustFromGraph(context.Background(), store, "onboarding", "invite_member")
	require.NoError(t, err)
	require.True(t, ok)
	require.Greater(t, assess.FlowIntegrity, 0.5)
	require.NotEmpty(t, assess.SecurityImpact)
}

func loadOnboardingStore(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
	require.NoError(t, err)
	body, err := os.ReadFile(filepath.Join("testdata", "knowledge-graph", "onboarding-flow", "graph.json"))
	require.NoError(t, err)
	graph, err := knowledge.ParseJSON(body)
	require.NoError(t, err)
	ctx := context.Background()
	for _, n := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, n))
	}
	for _, e := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, e))
	}
	require.NoError(t, store.Close())
	return repo
}
