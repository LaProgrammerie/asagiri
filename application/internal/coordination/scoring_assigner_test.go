package coordination_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestScoringAssignerPicksProfileByAssignment(t *testing.T) {
	assigner := &coordination.ScoringAssigner{
		Config: coordination.AssignerConfig{
			DefaultIsolation: coordination.IsolationIsolatedWorktree,
			Assignment: map[string]string{
				"implementation.high": "claude",
			},
			Profiles: map[string]coordination.AgentProfile{
				"claude-high": {Role: coordination.RoleImplementer, Agent: "claude", MaxContextTokens: 32000},
				"cursor-med":  {Role: coordination.RoleImplementer, Agent: "cursor", MaxContextTokens: 8000},
			},
		},
	}
	asg, err := assigner.Assign(context.Background(), executiongraph.GraphNode{
		ID:   "impl",
		Type: executiongraph.NodeTypeImplementation,
		Risk: executiongraph.RiskLevelHigh,
	})
	require.NoError(t, err)
	require.Equal(t, "claude", asg.AgentRef)
	require.Equal(t, "claude-high", asg.ProfileID)
}
