package checks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

func TestBlastRadiusRunnerFixture(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := BlastRadiusRunner{}.Run(context.Background(), Scope{
		TrustID:   "trust-br",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	require.NotNil(t, result.BlastRadius)
	require.GreaterOrEqual(t, result.BlastRadius.FlowsImpacted, 1)
	require.GreaterOrEqual(t, result.BlastRadius.CriticalAPIs, 1)
	require.NotEmpty(t, result.BlastRadius.MigrationRisk)
	require.NotEmpty(t, result.BlastRadius.PublicContractRisk)
}

func TestComputeBlastRadiusUnresolvedContract(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	pctx, skipped, err := loadProductContext(Scope{
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	require.False(t, skipped)
	br := computeBlastRadius(pctx.flow, pctx.bundle, pctx.bundleErr)
	require.Equal(t, "high", br.PublicContractRisk)
}

func TestComputeBlastRadiusBundleError(t *testing.T) {
	br := computeBlastRadius(product.Flow{ID: "f"}, analysis.Bundle{}, errors.New("bundle missing"))
	require.Equal(t, "unknown", br.MigrationRisk)
	require.Equal(t, "medium", br.PublicContractRisk)
}

func TestComputeBlastRadiusCountsCriticalAPIs(t *testing.T) {
	bundle := analysis.Bundle{
		Graphs: map[string]analysis.Graph{
			"api": {
				Nodes: []analysis.Node{
					{ID: "route:POST /a", Kind: "route", Name: "POST /a"},
					{ID: "route:GET /b", Kind: "route", Name: "GET /b"},
				},
			},
			"flow":       {},
			"dependency": {},
		},
	}
	br := computeBlastRadius(product.Flow{ID: "f"}, bundle, nil)
	require.Equal(t, 2, br.CriticalAPIs)
	require.Equal(t, "medium", br.PublicContractRisk)
}

func TestComputeBlastRadiusHighMigrationRisk(t *testing.T) {
	bundle := analysis.Bundle{
		Graphs: map[string]analysis.Graph{
			"flow": {
				Nodes: []analysis.Node{
					{ID: "flow:a"}, {ID: "flow:b"}, {ID: "flow:c"}, {ID: "flow:d"},
				},
			},
			"api":        {},
			"dependency": {Nodes: []analysis.Node{{ID: "pkg:a"}}},
		},
	}
	flow := product.Flow{
		ID:       "critical-flow",
		Business: product.FlowBusiness{Criticality: "high"},
	}
	br := computeBlastRadius(flow, bundle, nil)
	require.Equal(t, "high", br.MigrationRisk)
	require.GreaterOrEqual(t, br.FlowsImpacted, 4)
}

func TestBlastRadiusRunnerWarnsHighPublicContractRisk(t *testing.T) {
	repo := setupMinimalProductRepo(t)
	result, err := BlastRadiusRunner{}.Run(context.Background(), Scope{
		TrustID:   "trust-br-warn",
		Flow:      "workspace-onboarding",
		RepoRoot:  repo,
		ProductID: "minimal-product",
	}, testDeps(t, repo))
	require.NoError(t, err)
	requireFinding(t, result.Findings, "warning", "blast.radius", "public contract risk is high")
	require.NotNil(t, result.BlastRadius)
	require.Equal(t, "high", result.BlastRadius.PublicContractRisk)
}

func TestBlastRadiusRunnerSkippedNoProduct(t *testing.T) {
	result, err := BlastRadiusRunner{}.Run(context.Background(), Scope{
		TrustID: "trust-br-skip",
		Flow:    "workspace-onboarding",
	}, testDeps(t, t.TempDir()))
	require.NoError(t, err)
	require.Equal(t, statusSkipped, result.Status)
}
