package executiongraph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderFormatsDeterministic(t *testing.T) {
	body := readGoldenFixture(t, "parallel-independent")
	graph, err := ParseYAML(body)
	require.NoError(t, err)

	mermaid1, err := Render(graph, RenderFormatMermaid)
	require.NoError(t, err)
	mermaid2, err := Render(graph, RenderFormatMermaid)
	require.NoError(t, err)
	require.Equal(t, mermaid1, mermaid2)

	dot, err := Render(graph, RenderFormatDOT)
	require.NoError(t, err)
	require.Contains(t, dot, "digraph execution_graph")
	require.Contains(t, dot, `"implement-ui-onboarding"`)

	md, err := Render(graph, RenderFormatMarkdown)
	require.NoError(t, err)
	require.Contains(t, md, "# Execution Graph")
	require.Contains(t, md, "parallel_with")

	jsonOut, err := Render(graph, RenderFormatJSON)
	require.NoError(t, err)
	require.NotContains(t, jsonOut, `"graph_id"`)
	require.Contains(t, jsonOut, `"id": "graph-2026-05-27-b2c3d4e5"`)
}

func TestRenderPlanMD(t *testing.T) {
	body := readGoldenFixture(t, "simple-linear")
	graph, err := ParseYAML(body)
	require.NoError(t, err)

	plan := RenderPlanMD(graph)
	require.Contains(t, plan, "# Execution Plan")
	require.Contains(t, plan, "graph-2026-05-27-a1b2c3d4")
	require.Contains(t, plan, "```mermaid")
}

func TestRenderGoldenSnapshots(t *testing.T) {
	for _, scenario := range goldenScenarios() {
		t.Run(scenario+"/mermaid", func(t *testing.T) {
			graph := loadScenarioGraph(t, scenario)
			got, err := Render(graph, RenderFormatMermaid)
			require.NoError(t, err)
			assertGolden(t, scenario, "mermaid.txt", got)
		})
	}
}

func loadScenarioGraph(t *testing.T, scenario string) ExecutionGraph {
	t.Helper()
	graph, err := ParseYAML(readGoldenFixture(t, scenario))
	require.NoError(t, err)
	return graph
}

func assertGolden(t *testing.T, scenario, name, got string) {
	t.Helper()
	golden := filepath.Join("testdata", "execution-graph", scenario, name)
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
