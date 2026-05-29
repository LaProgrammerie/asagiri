package executiongraph

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoldenFixturesParseAndValidate(t *testing.T) {
	for _, scenario := range goldenScenarios() {
		t.Run(scenario, func(t *testing.T) {
			body := readGoldenFixture(t, scenario)
			graph, err := ParseYAML(body)
			require.NoError(t, err)
			require.NoError(t, graph.Validate())

			jsonBody, err := json.Marshal(graph)
			require.NoError(t, err)
			roundTrip, err := ParseJSON(jsonBody)
			require.NoError(t, err)
			require.NoError(t, roundTrip.Validate())
			require.Equal(t, graph.ID, roundTrip.ID)
		})
	}
}

func TestGoldenFixturesRepositoryRoundTrip(t *testing.T) {
	repo := NewRepository(t.TempDir())
	for _, scenario := range goldenScenarios() {
		t.Run(scenario, func(t *testing.T) {
			graph, err := ParseYAML(readGoldenFixture(t, scenario))
			require.NoError(t, err)
			_, _, err = repo.Save(graph)
			require.NoError(t, err)

			loaded, err := repo.Load(graph.ID)
			require.NoError(t, err)
			require.Equal(t, graph.ID, loaded.ID)
			require.Equal(t, len(graph.Nodes), len(loaded.Nodes))
			require.Equal(t, len(graph.Edges), len(loaded.Edges))
		})
	}
}

func TestStubInterfacesReturnNotImplemented(t *testing.T) {
	_, err := (StubPlanner{}).Build(t.Context(), GraphPlanRequest{})
	require.ErrorIs(t, err, ErrNotImplemented)
	_, err = (StubDependencyInferer{}).Infer(t.Context(), DependencyInput{})
	require.ErrorIs(t, err, ErrNotImplemented)
	_, err = (StubScheduler{}).Schedule(t.Context(), ScheduleRequest{})
	require.ErrorIs(t, err, ErrNotImplemented)
	_, err = (StubExecutor{}).Run(t.Context(), ExecutionGraph{})
	require.ErrorIs(t, err, ErrNotImplemented)
	_, err = (StubExecutor{}).Resume(t.Context(), "graph-2026-05-27-a1b2c3d4")
	require.ErrorIs(t, err, ErrNotImplemented)
}

func TestGoldenJSONSnapshots(t *testing.T) {
	for _, scenario := range goldenScenarios() {
		t.Run(scenario, func(t *testing.T) {
			graph := loadScenarioGraph(t, scenario)
			got, err := Render(graph, RenderFormatJSON)
			require.NoError(t, err)
			assertGolden(t, scenario, "execution-graph.json", got)
		})
	}
}
