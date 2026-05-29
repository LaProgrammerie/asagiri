package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func lot3Scope(repo string) Scope {
	return Scope{
		TrustID:   "trust-lot3",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
}

func TestPermissionsRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := PermissionsRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusPassed, result.Status)
}

func TestObservabilityRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := ObservabilityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusWarn, result.Status)
}

func TestSecurityRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := SecurityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusWarn, result.Status)
	requireFinding(t, result.Findings, "warning", "security.contract", "TODO:auth.signup")
}

func TestCostRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := CostRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusPassed, result.Status)
}

func TestAnalyticsRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := AnalyticsRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusPassed, result.Status)
}

func TestArchitectureRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := ArchitectureRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.NotEmpty(t, result.Evidence)
}

func TestPerformanceRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := PerformanceRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusPassed, result.Status)
}

func TestBackwardCompatibilityRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := BackwardCompatibilityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "compatibility.contract", "unresolved contract_ref")
}

func TestMigrationSafetyRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := MigrationSafetyRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.NotEqual(t, statusFailed, result.Status)
}
