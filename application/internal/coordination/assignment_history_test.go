package coordination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestMemoryAssignmentHistoryBoostsScoring(t *testing.T) {
	hist := &coordination.MemoryAssignmentHistory{}
	hist.RecordSuccess("claude", coordination.RoleImplementer)
	hist.RecordSuccess("claude", coordination.RoleImplementer)
	hist.RecordFailure("cursor", coordination.RoleImplementer)

	assigner := &coordination.ScoringAssigner{
		History: hist,
		Config: coordination.AssignerConfig{
			DefaultIsolation: coordination.IsolationIsolatedWorktree,
			Profiles: map[string]coordination.AgentProfile{
				"claude-impl": {Role: coordination.RoleImplementer, Agent: "claude"},
				"cursor-impl": {Role: coordination.RoleImplementer, Agent: "cursor"},
			},
		},
	}
	asg, err := assigner.Assign(t.Context(), executiongraph.GraphNode{
		ID:   "impl",
		Type: executiongraph.NodeTypeImplementation,
	})
	require.NoError(t, err)
	require.Equal(t, "claude", asg.AgentRef)
}
