package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractsRunnerWarnOnTODO(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	deps := testDeps(t, repo)
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := ContractsRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	require.Equal(t, statusWarn, result.Status)
	require.Less(t, result.Confidence, 1.0)

	var todoFinding *Finding
	for i := range result.Findings {
		f := &result.Findings[i]
		if f.Category == "architecture.contract" && f.Severity == "warning" {
			if todoFinding == nil {
				todoFinding = f
			}
		}
	}
	require.NotNil(t, todoFinding)
	require.Contains(t, todoFinding.Message, "step-2")
	require.Contains(t, todoFinding.Message, "TODO:auth.signup")
	require.NotEmpty(t, todoFinding.SuggestedFix)
}

func TestContractsRunnerResolvesPOSTInGraph(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	deps := testDeps(t, repo)
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := ContractsRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	for _, f := range result.Findings {
		if f.Category == "contract.openapi" && f.Severity == "warning" {
			require.NotContains(t, f.Message, "POST /api/workspaces")
		}
	}
}
