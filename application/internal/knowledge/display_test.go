package knowledge_test

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func TestFormatKnowledgeBuildUX(t *testing.T) {
	t.Parallel()
	out := knowledge.FormatKnowledgeBuild(knowledge.BuildResult{
		Nodes:         1284,
		Edges:         3912,
		Sources:       []string{"flows", "contracts", "code", "tests"},
		AvgConfidence: 0.82,
		StaleFiles:    0,
		Warnings: []string{
			"4 flow actions have no linked test",
			"2 API operations have no linked permission",
		},
	})
	require.Contains(t, out, "Asagiri Knowledge Graph")
	require.Contains(t, out, "Nodes:        1284")
	require.Contains(t, out, "Sources:      flows, contracts, code, tests")
	require.Contains(t, out, "Confidence:   0.82 avg")
	require.Contains(t, out, "Stale:        0")
	require.Contains(t, out, "Top warnings")
	require.True(t, strings.Contains(out, "no linked test"))
}
