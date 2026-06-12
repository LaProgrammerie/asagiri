package knowledge_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestQueryImplementsFromFixture(t *testing.T) {
	store, err := knowledge.OpenStore(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	graph := loadFixtureGraph(t, "onboarding-flow")
	ctx := context.Background()
	for _, n := range graph.Nodes {
		require.NoError(t, store.UpsertNode(ctx, n))
	}
	for _, e := range graph.Edges {
		require.NoError(t, store.UpsertEdge(ctx, e))
	}

	q := knowledge.NewQuerier(store)
	result, err := q.QueryImplements(ctx, "invite_member")
	require.NoError(t, err)
	ids := make(map[string]struct{})
	for _, n := range result.Nodes {
		ids[n.ID] = struct{}{}
	}
	require.Contains(t, ids, "action:invite_member")
	require.Contains(t, ids, "api_operation:POST_invitations")
	require.Contains(t, ids, "symbol:InvitationService_invite")
}
