package coordination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func validGraph() coordination.ExecutionGraph {
	return coordination.ExecutionGraph{
		ID:      "graph-2026-05-29-a1b2c3d4",
		Product: "workspace-saas",
		Flow:    "onboarding",
		Status:  executiongraph.GraphStatusPlanned,
		Strategy: executiongraph.Strategy{
			MaxParallel: 2,
			StopOnRisk:  executiongraph.RiskLevelHigh,
		},
		Nodes: []executiongraph.GraphNode{
			{ID: "n1", Type: executiongraph.NodeTypeInvestigation},
		},
	}
}

func TestPolicyEvaluatorOK(t *testing.T) {
	eval := coordination.PolicyEvaluator{
		Policies: coordination.CoordinationPolicies{MaxParallelAgents: 2},
	}
	result := eval.Evaluate(validGraph())
	require.True(t, result.OK)
	require.Empty(t, result.Errors)
}

func TestPolicyEvaluatorRejectsHighParallelism(t *testing.T) {
	g := validGraph()
	g.Strategy.MaxParallel = 5
	eval := coordination.PolicyEvaluator{
		Policies: coordination.CoordinationPolicies{MaxParallelAgents: 2},
	}
	result := eval.Evaluate(g)
	require.False(t, result.OK)
	require.NotEmpty(t, result.Errors)
}

func TestPolicyEvaluatorSecurityWarning(t *testing.T) {
	g := validGraph()
	g.Flow = "auth-login"
	eval := coordination.PolicyEvaluator{
		Policies: coordination.CoordinationPolicies{
			MaxParallelAgents:        2,
			RequireSecurityReviewFor: []string{"auth"},
		},
	}
	result := eval.Evaluate(g)
	require.True(t, result.OK)
	require.NotEmpty(t, result.Warnings)
}
