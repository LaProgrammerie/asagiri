package executiongraph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
)

func TestTrustGateDryRunSimulatesPass(t *testing.T) {
	runner := &DefaultRunner{}
	blocked, reason := runner.evaluateTrustGate(context.Background(), ExecutionGraph{Flow: "f", Product: "p"}, GraphNode{
		ID:   "trust-gate",
		Type: NodeTypeTrustVerification,
	}, true, RunOptions{DryRun: true})
	require.False(t, blocked)
	require.Empty(t, reason)
}

func TestTrustGateStrictWithoutConfigBlocks(t *testing.T) {
	runner := &DefaultRunner{RepoRoot: t.TempDir()}
	blocked, reason := runner.evaluateTrustGate(context.Background(), ExecutionGraph{Flow: "f", Product: "p"}, GraphNode{
		ID:   "trust-gate",
		Type: NodeTypeTrustVerification,
	}, true, RunOptions{StrictTrust: true})
	require.True(t, blocked)
	require.Contains(t, reason, "not configured")
}

func TestTrustGateStrictWithPassingEvaluator(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	runner := &DefaultRunner{RepoRoot: repo}
	gates := trust.NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {},
		},
	})
	eng := trust.NewEngineWithChecks(repo, checks.NewRegistry())
	eng.Gates = gates
	blocked, reason := runner.evaluateTrustGate(context.Background(), ExecutionGraph{Flow: "workspace-onboarding", Product: "minimal-product"}, GraphNode{
		ID:   "trust-gate",
		Type: NodeTypeTrustVerification,
	}, true, RunOptions{StrictTrust: true, Gates: gates, TrustEngine: eng})
	require.False(t, blocked)
	require.Empty(t, reason)
}

func TestTrustGateNonStrictSkipsEvaluation(t *testing.T) {
	runner := &DefaultRunner{}
	gates := trust.NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {RequiredChecks: []string{"contracts"}},
		},
	})
	blocked, _ := runner.evaluateTrustGate(context.Background(), ExecutionGraph{Flow: "f", Product: "p"}, GraphNode{
		ID:   "trust-gate",
		Type: NodeTypeTrustVerification,
	}, false, RunOptions{Gates: gates})
	require.False(t, blocked)
}

func TestPlannerBuildIncludesTrustAndInvestigationNodes(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	planner := Planner{RepoRoot: repo, Inferer: DefaultDependencyInferer{}}
	graph, err := planner.Build(context.Background(), GraphPlanRequest{
		Product:        "minimal-product",
		Flow:           "workspace-onboarding",
		IncludeReviews: true,
	})
	require.NoError(t, err)

	hasTrustGate := false
	hasInvestigation := false
	for _, n := range graph.Nodes {
		if n.Type == NodeTypeTrustVerification {
			hasTrustGate = true
		}
		if n.Type == NodeTypeInvestigation && n.ID == "investigate-invite-member" {
			hasInvestigation = true
		}
	}
	require.True(t, hasTrustGate)
	require.True(t, hasInvestigation)
}
