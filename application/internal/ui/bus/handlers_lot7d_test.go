package bus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRecommendedActionsRuntimeStopped(t *testing.T) {
	qb := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	res, err := qb.Query(context.Background(), GetRecommendedActionsQuery{})
	require.NoError(t, err)
	typed, ok := res.(RecommendedActionsResult)
	require.True(t, ok)
	require.NotEmpty(t, typed.Actions)
	require.Equal(t, "cmd.start-work", typed.Actions[0].ActionID)
}

func TestGetExplainWithGraphNodeContext(t *testing.T) {
	qb := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	res, err := qb.Query(context.Background(), GetExplainQuery{
		Context: ExplainContext{
			Focus: FocusContext{
				Kind:    FocusKindGraphNode,
				Subject: "implement",
				Detail:  "onboarding",
				Screen:  "graph",
			},
		},
	})
	require.NoError(t, err)
	typed, ok := res.(ExplainResult)
	require.True(t, ok)
	require.Equal(t, "Why is this node blocked?", typed.Question)
	require.NotEmpty(t, typed.SupportedQuestions)
	require.Contains(t, typed.Evidence[0], "Focus:")
}

func TestFormatContractRefPending(t *testing.T) {
	require.Equal(t, "pending: auth.signup", FormatContractRef("TODO:auth.signup"))
	require.Equal(t, "POST /login", FormatContractRef("POST /login"))
	require.Equal(t, "none", FormatContractRef(""))
}
