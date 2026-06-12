package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStaticAnalysisRunnerPass(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	deps := testDeps(t, repo)
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := StaticAnalysisRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	require.NotEqual(t, statusFailed, result.Status)
	require.Greater(t, result.Confidence, 0.0)
}
