package executiongraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCheckpointEvery(t *testing.T) {
	require.NoError(t, ValidateCheckpointEvery(""))
	require.NoError(t, ValidateCheckpointEvery("node"))
	require.NoError(t, ValidateCheckpointEvery("group"))
	require.Error(t, ValidateCheckpointEvery("daily"))
}

func TestShouldPersistCheckpointCadence(t *testing.T) {
	graph := ExecutionGraph{
		Checkpoints: []Checkpoint{{After: "implement-a"}},
	}
	require.True(t, ShouldPersistCheckpoint(graph, "implement-b", CheckpointEveryNode))
	require.False(t, ShouldPersistCheckpoint(graph, "implement-b", CheckpointEveryGroup))
	require.False(t, ShouldPersistCheckpoint(graph, "implement-b", ""))
	require.True(t, ShouldPersistCheckpoint(graph, "implement-a", ""))
}

func TestShouldPersistGroupCheckpointCadence(t *testing.T) {
	require.True(t, ShouldPersistGroupCheckpoint(CheckpointEveryGroup))
	require.False(t, ShouldPersistGroupCheckpoint(CheckpointEveryNode))
	require.False(t, ShouldPersistGroupCheckpoint(""))
}
