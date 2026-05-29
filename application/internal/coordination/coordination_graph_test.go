package coordination_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestPersistCoordinationGraph(t *testing.T) {
	repo := t.TempDir()
	graph := validGraph()
	assignments := []coordination.AgentAssignment{
		{NodeID: "n1", AgentRef: "local", Role: coordination.RoleInvestigator},
	}
	cg := coordination.BuildCoordinationGraph(graph, assignments)
	path, err := coordination.PersistCoordinationGraph(repo, cg)
	require.NoError(t, err)
	require.FileExists(t, path)

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var loaded coordination.CoordinationGraph
	require.NoError(t, json.Unmarshal(raw, &loaded))
	require.Equal(t, graph.ID, loaded.GraphID)
	require.NotEmpty(t, loaded.Links)

	_, err = coordination.PersistCoordinationGraph(repo, coordination.CoordinationGraph{GraphID: "../bad"})
	require.Error(t, err)
}

func TestBuildCoordinationGraphLinks(t *testing.T) {
	graph := validGraph()
	graph.Flow = "payments"
	graph.Nodes[0].Task = "investigate-task"
	cg := coordination.BuildCoordinationGraph(graph, nil)
	require.NotEmpty(t, cg.Links)
	foundFlow := false
	for _, l := range cg.Links {
		if l.Kind == coordination.LinkFlow {
			foundFlow = true
		}
	}
	require.True(t, foundFlow)
}
