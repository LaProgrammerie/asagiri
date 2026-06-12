package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFlowExtractorParsesActionsAndContractRef(t *testing.T) {
	repo := t.TempDir()
	product := "demo"
	flowDir := filepath.Join(repo, ".asagiri", "products", product, "flows")
	require.NoError(t, os.MkdirAll(flowDir, 0o755))

	doc := map[string]any{
		"id": "onboarding",
		"steps": []map[string]any{
			{"id": "s1", "action": "invite_member", "contract_ref": "POST /invitations"},
		},
	}
	body, err := yaml.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "onboarding.flow.yaml"), body, 0o644))

	nodes, edges, warnings, err := extractors.FlowExtractor{}.Extract(context.Background(), repo, product)
	require.NoError(t, err)
	require.Empty(t, warnings)

	ids := nodeIDs(nodes)
	require.Contains(t, ids, "flow:onboarding")
	require.Contains(t, ids, "action:invite_member")
	require.Contains(t, ids, "api_operation:POST_invitations")

	var hasActionRequires bool
	for _, e := range edges {
		if e.Type == knowledge.EdgeTypeRequires && e.From == "action:invite_member" && e.To == "api_operation:POST_invitations" {
			hasActionRequires = true
			require.GreaterOrEqual(t, e.Confidence, 0.7)
			require.LessOrEqual(t, e.Confidence, 0.95)
			require.Equal(t, "flows", e.Source.Extractor)
		}
	}
	require.True(t, hasActionRequires)
}

func nodeIDs(nodes []knowledge.GraphNode) map[string]struct{} {
	out := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		out[n.ID] = struct{}{}
	}
	return out
}
