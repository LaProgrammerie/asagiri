package investigation_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestResolveScopeFromGraphOnboarding(t *testing.T) {
	repo := loadKnowledgeFixture(t, "onboarding-flow")
	pack, err := investigation.ResolveScopeFromGraph(context.Background(), repo, investigation.GraphScopeOptions{
		UseKnowledgeGraph: true,
		Flow:              "onboarding",
		Action:            "invite_member",
	})
	require.NoError(t, err)
	require.Contains(t, pack.Files, "internal/invitation/invitation_service.go")
	require.Contains(t, pack.Tests, "internal/invitation/invitation_service_test.go")
	require.NotEmpty(t, pack.APIs)
	require.NotEmpty(t, pack.Events)
}

func TestAugmentRootCauseGraph(t *testing.T) {
	repo := loadKnowledgeFixture(t, "onboarding-flow")
	pack, err := investigation.ResolveScopeFromGraph(context.Background(), repo, investigation.GraphScopeOptions{
		UseKnowledgeGraph: true,
		Flow:              "onboarding",
		Action:            "invite_member",
	})
	require.NoError(t, err)
	rep := investigation.Report{
		Request: investigation.Request{Symptom: "invite fails"},
		Scope:   investigation.ResolvedScope{Flow: "onboarding", Action: "invite_member"},
	}
	g, err := investigation.BuildRootCauseGraphWithKnowledge(context.Background(), repo, rep, pack)
	require.NoError(t, err)
	require.NotEmpty(t, g.Nodes)
}

func TestEnrichFromKnowledgeGraph(t *testing.T) {
	repo := loadKnowledgeFixture(t, "onboarding-flow")
	scope := investigation.ResolvedScope{Flow: "onboarding", Action: "invite_member"}
	local := investigation.InvestigationResult{}
	pack, ok := investigation.EnrichFromKnowledgeGraph(context.Background(), repo, &scope, &local)
	require.True(t, ok)
	require.NotEmpty(t, pack.APIs)
	require.NotEmpty(t, local.CandidateFiles)
}

func loadKnowledgeFixture(t *testing.T, scenario string) string {
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
