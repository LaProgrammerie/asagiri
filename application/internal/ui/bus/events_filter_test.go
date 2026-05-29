package bus

import (
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestEventMatchesCategoryAndSearch(t *testing.T) {
	t.Parallel()

	ev := EventSummary{Type: "graph.generated", Source: "planner", ID: "evt-1"}
	require.True(t, eventMatchesCategory(ev, "all"))
	require.True(t, eventMatchesCategory(ev, "graph"))
	require.False(t, eventMatchesCategory(ev, "trust"))

	require.True(t, eventMatchesSearchText(ev, ""))
	require.True(t, eventMatchesSearchText(ev, "generated"))
	require.False(t, eventMatchesSearchText(ev, "trust"))
}

func TestEventFeedFilterTypesIncludeCoreCategories(t *testing.T) {
	t.Parallel()

	require.Contains(t, EventFeedFilterTypes, "runtime")
	require.Contains(t, EventFeedFilterTypes, "graph")
	require.Contains(t, EventFeedFilterTypes, "replay")
}

func TestMapRuntimeEventPreservesFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	got := mapRuntimeEvent(runtime.RuntimeEvent{
		ID:        "evt-1",
		Type:      "graph.generated",
		Source:    "planner",
		SessionID: "sess-1",
		CreatedAt: now,
	})
	require.Equal(t, "graph.generated", got.Type)
	require.Equal(t, "planner", got.Source)
	require.Equal(t, "sess-1", got.SessionID)
}
