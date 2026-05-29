package checks

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
)

func TestSecurityRunnerUnresolvedContractRef(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := SecurityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "security.contract", "unresolved security-sensitive contract: TODO:auth.signup")
}

func TestSecurityRunnerSensitiveWithoutAuth(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "no-auth.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: no-auth
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: delete_all
    sensitive: true
    errors:
      - forbidden
outcome: done
security:
  requires_authentication: false
`), 0o644))

	result, err := SecurityRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-sec", Flow: "no-auth", RepoRoot: repo, ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusFailed, result.Status)
	requireFinding(t, result.Findings, "error", "security.flow", "sensitive step s1 without auth requirement")
}

func TestSecurityRunnerHighCriticalityWithoutAuth(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "high-no-auth.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: high-no-auth
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: browse
    next: s2
  - id: s2
    screen: s2
    action: finish
outcome: done
business:
  criticality: high
metrics:
  - visits
security:
  requires_authentication: false
`), 0o644))

	result, err := SecurityRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-sec-auth", Flow: "high-no-auth", RepoRoot: repo, ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusWarn, result.Status)
	requireFinding(t, result.Findings, "warning", "security.flow", "high criticality flow without requires_authentication")
}

func TestSecurityRunnerMissingPermissionsContract(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	permPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "contracts", "permissions.yaml")
	require.NoError(t, os.Remove(permPath))

	result, err := SecurityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "security.contract", "permissions contract missing")
}

func TestSecurityRunnerSensitivePathsInScope(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	deps := testDeps(t, repo)
	deps.Investigate = func(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (investigation.InvestigationResult, error) {
		return investigation.InvestigationResult{
			SensitivePaths: []string{"application/internal/auth", "application/internal/billing"},
		}, nil
	}

	result, err := SecurityRunner{}.Run(context.Background(), lot3Scope(repo), deps)
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "security.contract", "2 sensitive paths in change scope")
}

func TestSecurityRunnerSkippedNoProduct(t *testing.T) {
	result, err := SecurityRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-sec-skip",
		Flow:    "workspace-onboarding",
	}, testDeps(t, t.TempDir()))
	require.NoError(t, err)
	require.Equal(t, statusSkipped, result.Status)
}
