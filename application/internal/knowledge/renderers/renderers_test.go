package renderers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/renderers"
	"github.com/stretchr/testify/require"
)

func TestRenderFormatsDeterministic(t *testing.T) {
	graph := loadFixtureGraph(t, "onboarding-flow")

	json1, err := renderers.Render(graph, renderers.FormatJSON)
	require.NoError(t, err)
	json2, err := renderers.Render(graph, renderers.FormatJSON)
	require.NoError(t, err)
	require.Equal(t, json1, json2)

	mermaid1, err := renderers.Render(graph, renderers.FormatMermaid)
	require.NoError(t, err)
	mermaid2, err := renderers.Render(graph, renderers.FormatMermaid)
	require.NoError(t, err)
	require.Equal(t, mermaid1, mermaid2)
}

func TestRenderGoldenSnapshots(t *testing.T) {
	for _, scenario := range []string{"minimal", "onboarding-flow"} {
		t.Run(scenario+"/json", func(t *testing.T) {
			graph := loadFixtureGraph(t, scenario)
			got, err := renderers.Render(graph, renderers.FormatJSON)
			require.NoError(t, err)
			assertGolden(t, scenario, "graph.json", got)
		})
		t.Run(scenario+"/dot", func(t *testing.T) {
			graph := loadFixtureGraph(t, scenario)
			got, err := renderers.Render(graph, renderers.FormatDOT)
			require.NoError(t, err)
			assertGolden(t, scenario, "graph.dot", got)
		})
		t.Run(scenario+"/mermaid", func(t *testing.T) {
			graph := loadFixtureGraph(t, scenario)
			got, err := renderers.Render(graph, renderers.FormatMermaid)
			require.NoError(t, err)
			assertGolden(t, scenario, "graph.mmd", got)
		})
	}
}

func loadFixtureGraph(t *testing.T, scenario string) knowledge.KnowledgeGraph {
	t.Helper()
	path := filepath.Join("..", "testdata", "knowledge-graph", scenario, "graph.json")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	graph, err := knowledge.ParseJSON(body)
	require.NoError(t, err)
	return graph
}

func assertGolden(t *testing.T, scenario, name, got string) {
	t.Helper()
	golden := filepath.Join("..", "testdata", "knowledge-graph", scenario, name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
	}
	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
		want = []byte(got)
	} else {
		require.NoError(t, err)
	}
	require.Equal(t, string(want), got)
}
