package coordination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestPoliciesFromConfigBridgesReviewPolicies(t *testing.T) {
	cfg := config.NewTestConfig("proj")
	cfg.Coordination.RequireIndependentReview = true
	cfg.Coordination.AllowSelfReview = false
	cfg.Coordination.DefaultIsolation = "readonly"
	cfg.Coordination.RequireSecurityReviewFor = []string{"auth", "payments"}

	eval := coordination.PoliciesFromConfig(cfg)
	require.Equal(t, cfg.Coordination.MaxParallelAgents, eval.Policies.MaxParallelAgents)
	require.True(t, eval.Policies.RequireIndependentReview)
	require.False(t, eval.Policies.AllowSelfReview)
	require.Equal(t, coordination.IsolationReadonly, eval.Policies.DefaultIsolation)
	require.Equal(t, []string{"auth", "payments"}, eval.Policies.RequireSecurityReviewFor)
}

func TestAssignerConfigFromConfigDefaultsIsolation(t *testing.T) {
	cfg := config.NewTestConfig("proj")
	ac := coordination.AssignerConfigFromConfig(cfg)
	require.Equal(t, coordination.IsolationIsolatedWorktree, ac.DefaultIsolation)
}

func TestAssignerConfigFromConfigNilDefaultsIsolation(t *testing.T) {
	ac := coordination.AssignerConfigFromConfig(nil)
	require.Equal(t, coordination.IsolationIsolatedWorktree, ac.DefaultIsolation)
}

func TestCoordinatorServicesFromConfig(t *testing.T) {
	cfg := config.NewTestConfig("proj")
	cfg.Coordination.Merge.Require = []string{"trust_passed"}
	cfg.Coordination.Merge.BlockIf = []string{"unresolved_conflicts"}
	svc := coordination.CoordinatorServicesFromConfig(cfg)
	require.NotNil(t, svc.Budget)
	require.NotNil(t, svc.Conflict)
	require.NotNil(t, svc.Escalation)
	require.NotNil(t, svc.Merge)
}

func TestMergeEvaluatorFromConfig(t *testing.T) {
	cfg := config.NewTestConfig("proj")
	cfg.Coordination.Merge.BlockIf = []string{"low_security_confidence"}
	eval := coordination.MergeEvaluatorFromConfig(cfg)
	require.Contains(t, eval.BlockIf, "low_security_confidence")
}
