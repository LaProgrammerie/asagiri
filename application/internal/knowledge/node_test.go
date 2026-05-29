package knowledge_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func TestNodeIDAndValidate(t *testing.T) {
	id := knowledge.NodeID(knowledge.NodeTypeFlow, "onboarding")
	require.Equal(t, "flow:onboarding", id)
	require.NoError(t, knowledge.ValidateNodeID(id))

	node := knowledge.GraphNode{
		ID:         id,
		Type:       knowledge.NodeTypeFlow,
		Name:       "onboarding",
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "graph.json"},
		Confidence: 1,
	}
	require.NoError(t, node.Validate())
}

func TestNodeValidateRejectsInvalidProvenance(t *testing.T) {
	node := knowledge.GraphNode{
		ID:         knowledge.NodeID(knowledge.NodeTypeFlow, "onboarding"),
		Type:       knowledge.NodeTypeFlow,
		Confidence: 1,
	}
	require.ErrorIs(t, node.Validate(), knowledge.ErrInvalidNode)
}

func TestInvalidNodeIDsFromFixture(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "knowledge-graph", "invalid-id", "ids.json"))
	require.NoError(t, err)
	var ids []string
	require.NoError(t, json.Unmarshal(body, &ids))
	for _, id := range ids {
		require.Error(t, knowledge.ValidateNodeID(id), "id %q should be invalid", id)
	}
}
