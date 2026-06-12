package executiongraph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlannerBuildMinimalProduct(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	planner := Planner{
		RepoRoot: repo,
		Inferer:  DefaultDependencyInferer{},
		Now: func() time.Time {
			return time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
		},
	}

	graph, err := planner.Build(t.Context(), GraphPlanRequest{
		Product:        "minimal-product",
		Flow:           "workspace-onboarding",
		FromProduct:    true,
		IncludeReviews: true,
	})
	require.NoError(t, err)
	require.NoError(t, graph.Validate())

	require.Equal(t, "minimal-product", graph.Product)
	require.Equal(t, "workspace-onboarding", graph.Flow)
	require.Equal(t, GraphStatusPlanned, graph.Status)
	require.NotEmpty(t, graph.ID)
	require.GreaterOrEqual(t, len(graph.Nodes), 5)
	require.NotEmpty(t, graph.Edges)

	nodeIDs := map[string]NodeType{}
	for _, n := range graph.Nodes {
		nodeIDs[n.ID] = n.Type
	}
	require.Equal(t, NodeTypeInvestigation, nodeIDs["investigate-onboarding"])
	require.Equal(t, NodeTypeImplementation, nodeIDs["implement-click-get-started"])
	require.Equal(t, NodeTypeImplementation, nodeIDs["implement-invite-member"])
	require.Equal(t, NodeTypeContractGeneration, nodeIDs["derive-contracts"])
	require.Equal(t, NodeTypeValidation, nodeIDs["verify-contracts"])
	require.Equal(t, NodeTypeReview, nodeIDs["security-review"])
	require.Equal(t, NodeTypeTrustVerification, nodeIDs["trust-gate"])

	requireContainsEdge(t, graph.Edges, GraphEdge{
		From: "derive-contracts",
		To:   "implement-invite-member",
		Type: EdgeTypeBlocks,
	})
	requireContainsEdge(t, graph.Edges, GraphEdge{
		From: "implement-invite-member",
		To:   "security-review",
		Type: EdgeTypeValidates,
	})
	requireContainsEdge(t, graph.Edges, GraphEdge{
		From: "verify-contracts",
		To:   "implement-click-get-started",
		Type: EdgeTypeRequires,
	})
}

func TestPlannerBuildDetectsCycle(t *testing.T) {
	planner := Planner{
		RepoRoot: t.TempDir(),
		Inferer:  cycleInferer{},
		Now:      func() time.Time { return time.Unix(0, 0).UTC() },
	}
	require.NoError(t, os.MkdirAll(filepath.Join(planner.RepoRoot, ".asagiri", "products", "p", "flows"), 0o755))
	flow := []byte(`id: f
title: F
entry_screen: s
steps:
  - id: s1
    screen: s
    action: a
  - id: s2
    screen: s
    action: b
`)
	require.NoError(t, os.WriteFile(filepath.Join(planner.RepoRoot, ".asagiri", "products", "p", "flows", "f.flow.yaml"), flow, 0o644))

	_, err := planner.Build(t.Context(), GraphPlanRequest{
		Product:        "p",
		Flow:           "f",
		IncludeReviews: false,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCycleDetected)
}

func TestPlannerBuildRequiresProductAndFlow(t *testing.T) {
	planner := NewPlanner(t.TempDir())
	_, err := planner.Build(t.Context(), GraphPlanRequest{})
	require.Error(t, err)
	_, err = planner.Build(t.Context(), GraphPlanRequest{Product: "p"})
	require.Error(t, err)
}

type cycleInferer struct{}

func (cycleInferer) Infer(_ context.Context, _ DependencyInput) ([]GraphEdge, error) {
	return []GraphEdge{
		{From: "implement-a", To: "implement-b", Type: EdgeTypeRequires},
		{From: "implement-b", To: "implement-a", Type: EdgeTypeRequires},
	}, nil
}
