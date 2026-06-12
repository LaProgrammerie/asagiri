package checks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlowsRunnerFailMissingEntry(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	badFlow := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "broken.flow.yaml")
	require.NoError(t, os.WriteFile(badFlow, []byte("id: broken\nsteps:\n  - id: s1\n    screen: s1\n    action: go\n"), 0o644))

	deps := testDeps(t, repo)
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "broken",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := FlowsRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	require.Equal(t, statusFailed, result.Status)
}

func TestFlowsRunnerPassMinimal(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	deps := testDeps(t, repo)
	scope := Scope{
		TrustID:   "trust-test",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}
	result, err := FlowsRunner{}.Run(context.Background(), scope, deps)
	require.NoError(t, err)
	require.NotEqual(t, statusFailed, result.Status)
}

func TestFlowsRunnerWarnMissingContractRef(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "gaps.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: gaps
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: go
outcome: done
observability: {}
`), 0o644))

	deps := testDeps(t, repo)
	result, err := FlowsRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-test", Flow: "gaps", RepoRoot: repo, ProductID: "minimal-product",
	}, deps)
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "flow.integrity", "missing contract_ref")
}

func TestFlowsRunnerFailSensitiveWithoutErrors(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "insecure.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: insecure
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: go
    sensitive: true
outcome: done
`), 0o644))

	result, err := FlowsRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-test", Flow: "insecure", RepoRoot: repo, ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusFailed, result.Status)
	requireFinding(t, result.Findings, "error", "flow.integrity", "sensitive action requires errors")
}

func TestFlowsRunnerFailHighCriticalityWithoutMetrics(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "high-no-metrics.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: high-no-metrics
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: go
    contract_ref: GET /x
outcome: done
business:
  criticality: high
`), 0o644))

	result, err := FlowsRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-test", Flow: "high-no-metrics", RepoRoot: repo, ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	require.Equal(t, statusFailed, result.Status)
	requireFinding(t, result.Findings, "error", "flow.integrity", "flow.metrics is required")
}

func TestFlowsRunnerInfoEmptyObservability(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "no-telemetry.flow.yaml")
	require.NoError(t, os.WriteFile(flowPath, []byte(`id: no-telemetry
entry_screen: s1
steps:
  - id: s1
    screen: s1
    action: go
    contract_ref: GET /x
outcome: done
observability: {}
`), 0o644))

	result, err := FlowsRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-test", Flow: "no-telemetry", RepoRoot: repo, ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	requireFinding(t, result.Findings, "info", "flow.observability", "observability block is empty")
}

func requireFinding(t *testing.T, findings []Finding, severity, category, msgSubstr string) {
	t.Helper()
	for _, f := range findings {
		if f.Severity == severity && f.Category == category && strings.Contains(f.Message, msgSubstr) {
			return
		}
	}
	t.Fatalf("finding %s/%s containing %q not found in %+v", severity, category, msgSubstr, findings)
}
