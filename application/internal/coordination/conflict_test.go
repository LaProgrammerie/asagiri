package coordination_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestDefaultConflictDetectorFileOverlap(t *testing.T) {
	graph := validGraph()
	graph.Nodes = []executiongraph.GraphNode{
		{ID: "a", Type: executiongraph.NodeTypeImplementation, Outputs: []string{"src/foo.go"}},
		{ID: "b", Type: executiongraph.NodeTypeReview, Outputs: []string{"src/foo.go"}},
	}
	d := coordination.DefaultConflictDetector{}
	conflicts, err := d.Detect(context.Background(), graph)
	require.NoError(t, err)
	require.NotEmpty(t, conflicts)
	require.Equal(t, coordination.ConflictFileOverlap, conflicts[0].Category)
}

func TestDefaultConflictDetectorContractDrift(t *testing.T) {
	graph := validGraph()
	graph.Nodes = []executiongraph.GraphNode{
		{ID: "c1", Type: executiongraph.NodeTypeContractGeneration, Outputs: []string{"api/a.yaml"}},
		{ID: "c2", Type: executiongraph.NodeTypeContractGeneration, Outputs: []string{"api/b.yaml"}},
	}
	conflicts, err := coordination.DefaultConflictDetector{}.Detect(context.Background(), graph)
	require.NoError(t, err)
	require.NotEmpty(t, conflicts)
	require.Equal(t, coordination.ConflictContractDrift, conflicts[0].Category)
}

func TestDefaultConflictDetectorTrustDowngrade(t *testing.T) {
	graph := validGraph()
	graph.Nodes = []executiongraph.GraphNode{
		{ID: "val", Type: executiongraph.NodeTypeValidation, Status: executiongraph.NodeStatusFailed},
		{ID: "trust", Type: executiongraph.NodeTypeTrustVerification},
	}
	graph.Edges = []executiongraph.GraphEdge{{From: "trust", To: "val", Type: executiongraph.EdgeTypeRequires}}
	conflicts, err := coordination.DefaultConflictDetector{}.Detect(context.Background(), graph)
	require.NoError(t, err)
	require.NotEmpty(t, conflicts)
	require.Equal(t, coordination.ConflictTrustDowngrade, conflicts[0].Category)
}
