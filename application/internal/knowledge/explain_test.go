package knowledge_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func TestExplainRejectsUnknownFlowActionLink(t *testing.T) {
	store := openOnboardingStore(t)
	q := knowledge.NewQuerier(store)

	_, err := q.ExplainShortestPath(context.Background(), knowledge.ExplainRequest{
		Flow:   "onboarding",
		Action: "unknown_action",
		Symbol: "InvitationService",
	})
	require.Error(t, err)
}
