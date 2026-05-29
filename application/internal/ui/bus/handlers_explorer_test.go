package bus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetGraphViewBlockedFiltersNodes(t *testing.T) {
	b := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	res, err := b.Query(context.Background(), GetGraphViewQuery{View: GraphViewBlocked})
	require.NoError(t, err)
	typed, ok := res.(GraphViewResult)
	require.True(t, ok)
	require.Equal(t, GraphViewBlocked, typed.View)
}

func TestGetGraphNodeDetailRequiresNode(t *testing.T) {
	b := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	_, err := b.Query(context.Background(), GetGraphNodeDetailQuery{})
	require.Error(t, err)
}

func TestExportGraphCommandCLIEquivalent(t *testing.T) {
	cmd := ExportGraphCommand{GraphID: "graph-001", Format: "json"}
	require.Contains(t, cmd.CLIEquivalent(), "asa graph visualize graph-001")
	require.Contains(t, cmd.CLIEquivalent(), "json")
}

func TestCompareReplayCommandCLIEquivalent(t *testing.T) {
	cmd := CompareReplayCommand{ReplayA: "a", ReplayB: "b"}
	require.Equal(t, "asa replay compare a b", cmd.CLIEquivalent())
}
