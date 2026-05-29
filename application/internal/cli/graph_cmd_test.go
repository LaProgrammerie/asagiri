package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

func TestGraphRunOptionsFromPersisted(t *testing.T) {
	graph := executiongraph.ExecutionGraph{
		Strategy: executiongraph.Strategy{
			StrictTrust:     true,
			CheckpointEvery: executiongraph.CheckpointEveryGroup,
		},
	}
	opts := graphRunOptionsFromPersisted(graph)
	require.True(t, opts.StrictTrust)
	require.Equal(t, executiongraph.CheckpointEveryGroup, opts.CheckpointEvery)
}
