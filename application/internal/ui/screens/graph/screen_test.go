package graph

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	now := time.Date(2026, 5, 29, 11, 0, 0, 0, time.UTC)
	got := Render(ViewModel{
		Graph: bus.GraphExplorerResult{
			GraphID: "graph-001",
			FlowID:  "onboarding",
			Status:  "running",
			Nodes: []bus.GraphNodeSummary{
				{ID: "investigate", Title: "investigate", Status: "succeeded", Risk: "low", CLIEquivalent: "asa graph status graph-001"},
				{ID: "implement", Title: "implement", Status: "blocked", Risk: "high", BlockedBy: []string{"investigate"}, CLIEquivalent: "asa graph status graph-001"},
			},
		},
		View: bus.GraphViewResult{
			GraphID: "graph-001",
			FlowID:  "onboarding",
			View:    bus.GraphViewTimeline,
			Nodes: []bus.GraphNodeSummary{
				{ID: "investigate", Title: "investigate", Status: "succeeded", Risk: "low"},
				{ID: "implement", Title: "implement", Status: "blocked", Risk: "high", BlockedBy: []string{"investigate"}},
			},
		},
		Events: []bus.EventSummary{
			{Type: "graph.generated", CreatedAt: now},
			{Type: "graph.blocked", CreatedAt: now.Add(time.Minute)},
		},
		Model:   NewModel(),
		ShowCLI: true,
	})

	golden := filepath.Join("testdata", "graph.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
	}
	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", golden)
	}
	require.NoError(t, err)
	require.Equal(t, string(want), got)
	require.NotContains(t, got, "(stub)")
}

func TestRenderEmptyGraphWithWarning(t *testing.T) {
	got := Render(ViewModel{
		Graph: bus.GraphExplorerResult{
			FlowID:  "onboarding",
			Warning: "flow graph unavailable",
		},
		ShowCLI: true,
	})
	require.Contains(t, got, "Graph: -")
	require.Contains(t, got, "Nodes\n- none")
	require.Contains(t, got, "Events\n")
}
