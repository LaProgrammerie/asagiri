package executiongraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssessNodeRiskSensitiveAction(t *testing.T) {
	node := GraphNode{
		ID:   "implement-invite-member",
		Type: NodeTypeImplementation,
	}
	binding := &TaskBinding{
		Action:    "invite_member",
		Sensitive: true,
	}
	edges := []GraphEdge{
		{From: "implement-invite-member", To: "security-review", Type: EdgeTypeValidates},
	}

	assessment := AssessNodeRisk(node, binding, edges)
	require.Equal(t, RiskLevelHigh, assessment.Level)
	require.Equal(t, 1, assessment.BlastRadius)
	require.Contains(t, assessment.RequiredChecks, "tests")
	require.Contains(t, assessment.RequiredChecks, "security")
	require.Contains(t, assessment.Reasons, "touches sensitive action")
}

func TestAssessNodeRiskPermissionModelCritical(t *testing.T) {
	node := GraphNode{ID: "implement-permission", Type: NodeTypeImplementation}
	binding := &TaskBinding{
		Action:    "update_permission_model",
		Sensitive: true,
	}

	assessment := AssessNodeRisk(node, binding, nil)
	require.Equal(t, RiskLevelCritical, assessment.Level)
	require.True(t, assessment.RequiresHumanApproval)
}

func TestApplyRiskEnrichmentAddsManualApproval(t *testing.T) {
	nodes := []GraphNode{
		{ID: "trust-gate", Type: NodeTypeTrustVerification, Risk: RiskLevelHigh},
		{
			ID:                    "implement-invite-member",
			Type:                  NodeTypeImplementation,
			Risk:                  RiskLevelCritical,
			RequiresHumanApproval: true,
		},
	}
	bindings := []TaskBinding{{
		NodeID:    "implement-invite-member",
		Action:    "update_permission_model",
		Sensitive: true,
	}}

	updated, edges := ApplyRiskEnrichment(nodes, bindings, nil, nil)
	require.Len(t, updated, 3)
	require.Equal(t, NodeTypeManualApproval, updated[2].Type)
	require.Equal(t, "manual-approval", updated[2].ID)
	require.Len(t, edges, 1)
	require.Equal(t, EdgeTypeRequiresHumanApproval, edges[0].Type)
}

func TestHighestRisk(t *testing.T) {
	nodes := []GraphNode{
		{Risk: RiskLevelLow},
		{Risk: RiskLevelHigh},
		{Risk: RiskLevelMedium},
	}
	require.Equal(t, RiskLevelHigh, HighestRisk(nodes))
}

func TestDefaultAgentForImplementationRisk(t *testing.T) {
	require.Equal(t, "dev", DefaultAgentFor(GraphNode{
		Type: NodeTypeImplementation,
		Risk: RiskLevelMedium,
	}, nil))
	require.Equal(t, "claude", DefaultAgentFor(GraphNode{
		Type: NodeTypeImplementation,
		Risk: RiskLevelHigh,
	}, nil))
	require.Equal(t, "reviewer", DefaultAgentFor(GraphNode{
		Type: NodeTypeReview,
		Risk: RiskLevelHigh,
	}, nil))
}
