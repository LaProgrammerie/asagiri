package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryRunAllFourteenChecks(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	reg := NewDefaultRegistry(testDeps(t, repo))
	out, err := reg.RunAll(context.Background(), Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	})
	require.NoError(t, err)
	require.Len(t, out, 15)
	var blast *BlastRadiusSummary
	for _, c := range out {
		if c.Type == typeBlastRadius {
			blast = c.BlastRadius
			break
		}
	}
	require.NotNil(t, blast)
	require.GreaterOrEqual(t, blast.FlowsImpacted, 1)
	require.NotEmpty(t, blast.MigrationRisk)
}
