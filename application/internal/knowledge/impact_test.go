package knowledge_test

import (
	"context"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestGraphImpactAnalyzerByFlowAction(t *testing.T) {
	store := openOnboardingStore(t)
	analyzer := knowledge.NewImpactAnalyzer(store)

	result, err := analyzer.Analyze(context.Background(), knowledge.ImpactRequest{
		Flow:   "onboarding",
		Action: "invite_member",
	})
	require.NoError(t, err)
	require.Equal(t, "onboarding / invite_member", result.Input)
	require.Contains(t, result.ImpactedFlows, "onboarding / invite_member")
	require.Contains(t, result.ImpactedAPIs, "POST /invitations")
	require.Contains(t, result.ImpactedEvents, "member.invited")
	require.Contains(t, result.ImpactedTests, "InvitationServiceTest")
	require.NotEmpty(t, result.Risk)
	require.NotEmpty(t, result.RecommendedChecks)

	out := knowledge.FormatImpactAnalysis(result)
	require.Contains(t, out, "Impact Analysis")
	require.Contains(t, out, "Impacted flows:")
	require.Contains(t, out, "Risk:")
}

func TestGraphImpactAnalyzerByFile(t *testing.T) {
	store := openOnboardingStore(t)
	analyzer := knowledge.NewImpactAnalyzer(store)

	result, err := analyzer.Analyze(context.Background(), knowledge.ImpactRequest{
		File: "internal/invitation/invitation_service.go",
	})
	require.NoError(t, err)
	require.Equal(t, "internal/invitation/invitation_service.go", result.Input)
	require.Contains(t, result.ImpactedFlows, "onboarding / invite_member")
	require.Contains(t, result.ImpactedAPIs, "POST /invitations")
	require.Contains(t, result.ImpactedEvents, "member.invited")
	require.Contains(t, result.ImpactedTests, "InvitationServiceTest")
}

func TestGraphImpactAnalyzerRejectsMixedTargets(t *testing.T) {
	store := openOnboardingStore(t)
	analyzer := knowledge.NewImpactAnalyzer(store)

	_, err := analyzer.Analyze(context.Background(), knowledge.ImpactRequest{
		File:   "internal/invitation/invitation_service.go",
		Flow:   "onboarding",
		Action: "invite_member",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not both")
}

func TestExplainShortestPathOnboarding(t *testing.T) {
	store := openOnboardingStore(t)
	q := knowledge.NewQuerier(store)

	result, err := q.ExplainShortestPath(context.Background(), knowledge.ExplainRequest{
		Flow:   "onboarding",
		Action: "invite_member",
		Symbol: "InvitationService",
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(result.Steps), 2)
	require.Equal(t, "action:invite_member", result.Steps[0].Node.ID)
	last := result.Steps[len(result.Steps)-1].Node
	require.Equal(t, "symbol:InvitationService_invite", last.ID)

	out := knowledge.FormatKnowledgeExplain(result)
	require.Contains(t, out, "Knowledge path")
	require.True(t, strings.Contains(out, "invite_member"))
	require.True(t, strings.Contains(out, "InvitationService"))
}

func openOnboardingStore(t *testing.T) knowledge.GraphStore {
	t.Helper()
	repo := t.TempDir()
	store, err := knowledge.OpenStore(repo)
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
	return store
}
