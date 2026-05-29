package executiongraph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/stretchr/testify/require"
)

func TestDetectCyclesNoCycle(t *testing.T) {
	nodes := []GraphNode{
		{ID: "a", Type: NodeTypeInvestigation},
		{ID: "b", Type: NodeTypeImplementation},
		{ID: "c", Type: NodeTypeValidation},
	}
	edges := []GraphEdge{
		{From: "a", To: "b", Type: EdgeTypeRequires},
		{From: "b", To: "c", Type: EdgeTypeMustRunAfter},
		{From: "a", To: "c", Type: EdgeTypeProducesContextFor},
	}
	require.NoError(t, DetectCycles(nodes, edges))
}

func TestDetectCyclesFindsCycle(t *testing.T) {
	nodes := []GraphNode{
		{ID: "a", Type: NodeTypeImplementation},
		{ID: "b", Type: NodeTypeImplementation},
		{ID: "c", Type: NodeTypeImplementation},
	}
	edges := []GraphEdge{
		{From: "a", To: "b", Type: EdgeTypeRequires},
		{From: "b", To: "c", Type: EdgeTypeMustRunAfter},
		{From: "c", To: "a", Type: EdgeTypeBlocks},
	}
	err := DetectCycles(nodes, edges)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCycleDetected)
}

func TestDetectCyclesIgnoresNonCycleEdgeTypes(t *testing.T) {
	nodes := []GraphNode{
		{ID: "a", Type: NodeTypeImplementation},
		{ID: "b", Type: NodeTypeImplementation},
	}
	edges := []GraphEdge{
		{From: "a", To: "b", Type: EdgeTypeValidates},
		{From: "b", To: "a", Type: EdgeTypeParallelWith},
	}
	require.NoError(t, DetectCycles(nodes, edges))
}

func TestDetectCyclesRollbackDependsOn(t *testing.T) {
	nodes := []GraphNode{
		{ID: "deploy", Type: NodeTypeImplementation},
		{ID: "migrate", Type: NodeTypeImplementation},
		{ID: "rollback", Type: NodeTypeRollback},
	}
	edges := []GraphEdge{
		{From: "deploy", To: "migrate", Type: EdgeTypeRequires},
		{From: "migrate", To: "rollback", Type: EdgeTypeRollbackDependsOn},
		{From: "rollback", To: "deploy", Type: EdgeTypeMustRunAfter},
	}
	err := DetectCycles(nodes, edges)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCycleDetected)
}

func TestDetectCyclesReportsCyclePath(t *testing.T) {
	nodes := []GraphNode{
		{ID: "a", Type: NodeTypeImplementation},
		{ID: "b", Type: NodeTypeImplementation},
		{ID: "c", Type: NodeTypeImplementation},
	}
	edges := []GraphEdge{
		{From: "a", To: "b", Type: EdgeTypeRequires},
		{From: "b", To: "c", Type: EdgeTypeRequires},
		{From: "c", To: "a", Type: EdgeTypeBlocks},
	}
	err := DetectCycles(nodes, edges)
	require.Error(t, err)
	require.Contains(t, err.Error(), "a")
	require.Contains(t, err.Error(), "b")
	require.Contains(t, err.Error(), "c")
}

func TestInferProductRequired(t *testing.T) {
	inferer := DefaultDependencyInferer{}
	_, err := inferer.Infer(t.Context(), DependencyInput{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "product required")
}

func TestInferFlowStepOrderRequires(t *testing.T) {
	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product: "workspace-saas",
		Nodes: []GraphNode{
			{ID: "implement-click-get-started", Type: NodeTypeImplementation},
			{ID: "implement-invite-member", Type: NodeTypeImplementation},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-click-get-started",
				StepIndex:   0,
				Action:      "click_get_started",
				ContractRef: "POST /api/workspaces",
			},
			{
				NodeID:      "implement-invite-member",
				StepIndex:   1,
				Action:      "invite_member",
				ContractRef: "POST /api/members",
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "implement-click-get-started",
		To:   "implement-invite-member",
		Type: EdgeTypeRequires,
	})
}

func TestInferMissingPermissionBlocksImplementation(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product:  "minimal-product",
		Flow:     "workspace-onboarding",
		RepoRoot: repo,
		Nodes: []GraphNode{
			{ID: "derive-contracts", Type: NodeTypeContractGeneration},
			{ID: "implement-invite-member", Type: NodeTypeImplementation},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-invite-member",
				StepIndex:   1,
				Action:      "invite_member",
				ContractRef: "POST /api/members",
				Sensitive:   true,
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "derive-contracts",
		To:   "implement-invite-member",
		Type: EdgeTypeBlocks,
	})
}

func TestInferSensitiveActionFromFlowYAML(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	flowPath := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows", "workspace-onboarding.flow.yaml")
	raw, err := os.ReadFile(flowPath)
	require.NoError(t, err)
	raw = []byte(strings.Replace(string(raw),
		"  requires_authentication: true",
		"  requires_authentication: true\n  sensitive_actions:\n    - invite_member",
		1))
	require.NoError(t, os.WriteFile(flowPath, raw, 0o644))

	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product:  "minimal-product",
		Flow:     "workspace-onboarding",
		RepoRoot: repo,
		Nodes: []GraphNode{
			{ID: "implement-invite-member", Type: NodeTypeImplementation},
			{ID: "security-review", Type: NodeTypeReview},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-invite-member",
				StepIndex:   1,
				Action:      "invite_member",
				ContractRef: "POST /api/members",
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "implement-invite-member",
		To:   "security-review",
		Type: EdgeTypeValidates,
	})
}

func TestInferAPIRequiresChain(t *testing.T) {
	bundle := analysis.Bundle{
		Product: "workspace-saas",
		Graphs: map[string]analysis.Graph{
			"api": {
				Kind: "api",
				Nodes: []analysis.Node{
					{ID: "route:POST /api/workspaces", Kind: "route", Name: "POST /api/workspaces"},
				},
			},
		},
	}
	inferer := DefaultDependencyInferer{
		LoadBundle: func(_, _ string) (analysis.Bundle, error) { return bundle, nil },
	}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product: "workspace-saas",
		Flow:    "workspace-onboarding",
		Nodes: []GraphNode{
			{ID: "implement-click-get-started", Type: NodeTypeImplementation},
			{ID: "implement-invite-member", Type: NodeTypeImplementation},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-click-get-started",
				StepIndex:   0,
				Action:      "click_get_started",
				ContractRef: "POST /api/workspaces",
			},
			{
				NodeID:      "implement-invite-member",
				StepIndex:   1,
				Action:      "invite_member",
				ContractRef: "POST /api/workspaces",
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "implement-click-get-started",
		To:   "implement-invite-member",
		Type: EdgeTypeRequires,
	})
}

func TestInferSharedFileMustRunAfter(t *testing.T) {
	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product: "workspace-saas",
		Nodes: []GraphNode{
			{ID: "implement-a", Type: NodeTypeImplementation},
			{ID: "implement-b", Type: NodeTypeImplementation},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:     "implement-a",
				StepIndex:  0,
				Action:     "a",
				ScopePaths: []string{"application/internal/foo/**"},
			},
			{
				NodeID:     "implement-b",
				StepIndex:  1,
				Action:     "b",
				ScopePaths: []string{"application/internal/foo/bar.go"},
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "implement-a",
		To:   "implement-b",
		Type: EdgeTypeMustRunAfter,
	})
}

func TestInferSecurityReviewChain(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product:  "minimal-product",
		Flow:     "workspace-onboarding",
		RepoRoot: repo,
		Nodes: []GraphNode{
			{ID: "implement-invite-member", Type: NodeTypeImplementation},
			{ID: "security-review", Type: NodeTypeReview},
			{ID: "trust-gate", Type: NodeTypeTrustVerification},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-invite-member",
				StepIndex:   1,
				Action:      "invite_member",
				ContractRef: "TODO:auth.signup",
				Sensitive:   true,
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "implement-invite-member",
		To:   "security-review",
		Type: EdgeTypeValidates,
	})
	requireContainsEdge(t, edges, GraphEdge{
		From: "security-review",
		To:   "trust-gate",
		Type: EdgeTypeRequires,
	})
}

func TestInferBackwardCompatRequires(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product:  "minimal-product",
		Flow:     "workspace-onboarding",
		RepoRoot: repo,
		Nodes: []GraphNode{
			{ID: "derive-contracts", Type: NodeTypeContractGeneration},
			{ID: "implement-click-get-started", Type: NodeTypeImplementation},
			{ID: "verify-contracts", Type: NodeTypeValidation},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-click-get-started",
				StepIndex:   0,
				Action:      "click_get_started",
				ContractRef: "POST /api/workspaces",
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "derive-contracts",
		To:   "verify-contracts",
		Type: EdgeTypeValidates,
	})
	requireContainsEdge(t, edges, GraphEdge{
		From: "verify-contracts",
		To:   "implement-click-get-started",
		Type: EdgeTypeRequires,
	})
}

func TestInferContractBlocksImplementation(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	inferer := DefaultDependencyInferer{}
	edges, err := inferer.Infer(t.Context(), DependencyInput{
		Product:  "minimal-product",
		Flow:     "workspace-onboarding",
		RepoRoot: repo,
		Nodes: []GraphNode{
			{ID: "derive-contracts", Type: NodeTypeContractGeneration},
			{ID: "implement-invite-member", Type: NodeTypeImplementation},
		},
		TaskBindings: []TaskBinding{
			{
				NodeID:      "implement-invite-member",
				StepIndex:   1,
				Action:      "invite_member",
				ContractRef: "TODO:auth.signup",
				Sensitive:   true,
			},
		},
	})
	require.NoError(t, err)
	requireContainsEdge(t, edges, GraphEdge{
		From: "derive-contracts",
		To:   "implement-invite-member",
		Type: EdgeTypeBlocks,
	})
}

func requireContainsEdge(t *testing.T, edges []GraphEdge, want GraphEdge) {
	t.Helper()
	for _, e := range edges {
		if e.From == want.From && e.To == want.To && e.Type == want.Type {
			return
		}
	}
	t.Fatalf("edge not found: %+v in %v", want, edges)
}

func writeMinimalPlanningFixture(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "trust", "checks", "testdata", "minimal-product")
	repo := t.TempDir()

	copyTree(t, src, filepath.Join(repo, ".asagiri", "products", "minimal-product"))

	bundleRaw, err := os.ReadFile(filepath.Join("..", "trust", "checks", "testdata", "graphs-minimal.json"))
	require.NoError(t, err)
	dir := filepath.Join(repo, ".asagiri", "analysis", "minimal-product")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "graphs.json"), bundleRaw, 0o644))

	return repo
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	require.NoError(t, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	}))
}

func TestLoadBundleUsedByInferer(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	var loaded bool
	inferer := DefaultDependencyInferer{
		LoadBundle: func(root, productID string) (analysis.Bundle, error) {
			loaded = true
			require.Equal(t, repo, root)
			require.Equal(t, "minimal-product", productID)
			return analysis.LoadBundle(root, productID)
		},
	}
	_, err := inferer.Infer(t.Context(), DependencyInput{
		Product:  "minimal-product",
		Flow:     "workspace-onboarding",
		RepoRoot: repo,
		Nodes:    []GraphNode{{ID: "implement-click-get-started", Type: NodeTypeImplementation}},
		TaskBindings: []TaskBinding{{
			NodeID:      "implement-click-get-started",
			ContractRef: "POST /api/workspaces",
		}},
	})
	require.NoError(t, err)
	require.True(t, loaded)
}

func TestBundleJSONFixtureValid(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "trust", "checks", "testdata", "graphs-minimal.json"))
	require.NoError(t, err)
	var bundle analysis.Bundle
	require.NoError(t, json.Unmarshal(raw, &bundle))
	require.Contains(t, bundle.Graphs, "api")
}
