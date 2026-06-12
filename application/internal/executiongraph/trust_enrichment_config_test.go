package executiongraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyTrustEnrichmentTrustRequiredForHighRisk(t *testing.T) {
	nodes := []GraphNode{
		{ID: "implement-invite-member", Type: NodeTypeImplementation, Title: "Invite", Agent: "cursor", Risk: RiskLevelHigh},
	}
	bindings := []TaskBinding{{
		NodeID:    "implement-invite-member",
		Action:    "invite_member",
		Sensitive: true,
	}}
	out, edges := ApplyTrustEnrichment(nodes, bindings, nil, TrustEnrichmentInput{
		Gates: TrustGateConfig{TrustRequiredForHighRisk: true},
	})
	require.Len(t, out, 2)
	require.Equal(t, NodeTypeTrustVerification, out[1].Type)
	require.Equal(t, "trust-gate-invite-member", out[1].ID)
	require.Len(t, edges, 1)
}

func TestApplyRiskEnrichmentHumanApprovalForPublicContract(t *testing.T) {
	nodes := []GraphNode{
		{ID: "implement-api", Type: NodeTypeImplementation, Title: "API", Agent: "cursor"},
	}
	bindings := []TaskBinding{{
		NodeID:      "implement-api",
		Action:      "publish_api",
		ContractRef: "POST /v1/widgets",
	}}
	out, edges := ApplyRiskEnrichment(nodes, bindings, nil, []string{"public_contract_change"})
	require.True(t, out[0].RequiresHumanApproval)
	require.Len(t, edges, 1)
	require.Equal(t, NodeTypeManualApproval, out[1].Type)
}
