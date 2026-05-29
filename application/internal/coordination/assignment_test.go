package coordination_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestRoleForNodeType(t *testing.T) {
	cases := []struct {
		nodeType executiongraph.NodeType
		want     coordination.AgentRole
	}{
		{executiongraph.NodeTypeInvestigation, coordination.RoleInvestigator},
		{executiongraph.NodeTypeArchitectureDerivation, coordination.RoleArchitect},
		{executiongraph.NodeTypeImplementation, coordination.RoleImplementer},
		{executiongraph.NodeTypeReview, coordination.RoleReviewer},
		{executiongraph.NodeTypeValidation, coordination.RoleValidator},
		{executiongraph.NodeTypeDocumentation, coordination.RoleDocumenter},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, coordination.RoleForNodeType(tc.nodeType))
	}
}

func TestValidateRoleAndIsolation(t *testing.T) {
	require.NoError(t, coordination.ValidateRole(coordination.RoleReviewer))
	require.Error(t, coordination.ValidateRole(coordination.AgentRole("unknown")))

	require.NoError(t, coordination.ValidateIsolation(coordination.IsolationIsolatedWorktree))
	require.Error(t, coordination.ValidateIsolation(coordination.IsolationMode("bad")))
}

func TestDefaultAssignerDelegatesAndOverrides(t *testing.T) {
	assigner := &coordination.DefaultAssigner{
		Config: coordination.AssignerConfig{
			DefaultIsolation: coordination.IsolationIsolatedWorktree,
			Assignment: map[string]string{
				"investigation":        "local",
				"implementation.high":  "claude",
				"implementation.medium": "cursor",
			},
			Profiles: map[string]coordination.AgentProfile{
				"codex-reviewer": {
					Role:      coordination.RoleReviewer,
					Agent:     "codex",
					Isolation: coordination.IsolationReadonly,
				},
			},
		},
	}

	inv, err := assigner.Assign(context.Background(), executiongraph.GraphNode{
		ID:   "n1",
		Type: executiongraph.NodeTypeInvestigation,
	})
	require.NoError(t, err)
	require.Equal(t, "local", inv.AgentRef)
	require.Equal(t, coordination.RoleInvestigator, inv.Role)
	require.Equal(t, coordination.IsolationIsolatedWorktree, inv.Isolation)

	impl, err := assigner.Assign(context.Background(), executiongraph.GraphNode{
		ID:   "n2",
		Type: executiongraph.NodeTypeImplementation,
		Risk: executiongraph.RiskLevelHigh,
	})
	require.NoError(t, err)
	require.Equal(t, "claude", impl.AgentRef)
	require.Equal(t, coordination.RoleImplementer, impl.Role)

	review, err := assigner.Assign(context.Background(), executiongraph.GraphNode{
		ID:   "n3",
		Type: executiongraph.NodeTypeReview,
		Risk: executiongraph.RiskLevelHigh,
	})
	require.NoError(t, err)
	require.Equal(t, "codex", review.AgentRef)
	require.Equal(t, coordination.RoleReviewer, review.Role)
	require.Equal(t, coordination.IsolationReadonly, review.Isolation)
	require.Equal(t, "codex-reviewer", review.ProfileID)
}
