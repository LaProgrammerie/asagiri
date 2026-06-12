package graphcli

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
	opts := GraphRunOptionsFromPersisted(graph)
	require.True(t, opts.StrictTrust)
	require.Equal(t, executiongraph.CheckpointEveryGroup, opts.CheckpointEvery)
}

func TestRootCommand_Subcommands(t *testing.T) {
	cmd := RootCommand(Options{DryRun: nil, RunRootUI: nil})
	names := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	require.ElementsMatch(t, []string{"run", "status", "resume", "rollback", "visualize"}, names)
}
