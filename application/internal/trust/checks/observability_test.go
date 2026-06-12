package checks

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObservabilityRunnerMissingContract(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	obsPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "contracts", "observability.yaml")
	require.NoError(t, os.Remove(obsPath))

	result, err := ObservabilityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusFailed, result.Status)
	requireFinding(t, result.Findings, "error", "observability.contract", "observability.yaml missing")
}

func TestObservabilityRunnerEmptyFlowObservability(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "no-telemetry.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: no-telemetry
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: go
    next: s2
  - id: s2
    screen: s2
    action: finish
outcome: done
business:
  criticality: high
metrics:
  - step_count
observability: {}
`), 0o644))

	result, err := ObservabilityRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-obs", Flow: "no-telemetry", RepoRoot: repo, ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "observability.flow", "flow observability block is empty")
}

func TestObservabilityRunnerSensitiveStepMissingFailureMetric(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := ObservabilityRunner{}.Run(context.Background(), lot3Scope(repo), testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusWarn, result.Status)
	requireFinding(t, result.Findings, "warning", "observability.flow", "invite_member failure metric missing")
}

func TestObservabilityRunnerSkippedNoProduct(t *testing.T) {
	result, err := ObservabilityRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-obs-skip",
		Flow:    "workspace-onboarding",
	}, testDeps(t, t.TempDir()))
	require.NoError(t, err)
	require.Equal(t, statusSkipped, result.Status)
}
