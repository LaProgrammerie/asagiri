package components

import (
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderEventFeedShowsStubControlsAndEvents(t *testing.T) {
	now := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	got := RenderEventFeed(EventFeedViewModel{
		Events: []bus.EventSummary{
			{Type: "graph.generated", CreatedAt: now},
			{Type: "trust.completed", CreatedAt: now.Add(time.Minute)},
		},
		Limit: 2,
	})
	require.Contains(t, got, "Filter: all (stub)")
	require.Contains(t, got, "Search: (none)")
	require.Contains(t, got, "10:00:00  graph.generated")
	require.Contains(t, got, "10:01:00  trust.completed")
}

func TestRenderEventFeedAppliesFilterAndSearch(t *testing.T) {
	now := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	got := RenderEventFeed(EventFeedViewModel{
		Events: []bus.EventSummary{
			{Type: "graph.generated", CreatedAt: now},
			{Type: "trust.completed", CreatedAt: now.Add(time.Minute)},
		},
		Filter: "graph",
		Search: "generated",
		Limit:  5,
	})
	require.Contains(t, got, "graph.generated")
	require.NotContains(t, got, "trust.completed")
}

