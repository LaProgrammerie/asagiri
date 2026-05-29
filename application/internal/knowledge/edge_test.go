package knowledge_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func TestEdgeIDAndValidate(t *testing.T) {
	from := knowledge.NodeID(knowledge.NodeTypeAction, "invite_member")
	to := knowledge.NodeID(knowledge.NodeTypeAPIOperation, "POST_invitations")
	id := knowledge.EdgeID(knowledge.EdgeTypeRequires, from, to)
	require.Equal(t, "requires:action_invite_member>api_operation_POST_invitations", id)

	edge := knowledge.GraphEdge{
		ID:         id,
		From:       from,
		To:         to,
		Type:       knowledge.EdgeTypeRequires,
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "graph.json"},
		Confidence: 0.9,
	}
	require.NoError(t, edge.Validate())
}

func TestEdgeValidateRejectsMismatchedTypePrefix(t *testing.T) {
	edge := knowledge.GraphEdge{
		ID:         knowledge.EdgeID(knowledge.EdgeTypeCalls, "flow:onboarding", "action:invite_member"),
		From:       "flow:onboarding",
		To:         "action:invite_member",
		Type:       knowledge.EdgeTypeRequires,
		Source:     knowledge.GraphSource{Kind: "fixture"},
		Confidence: 0.9,
	}
	require.ErrorIs(t, edge.Validate(), knowledge.ErrInvalidEdge)
}
