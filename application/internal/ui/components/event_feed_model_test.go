package components

import (
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestEventFeedModelSearchModeFiltersSelection(t *testing.T) {
	m := NewEventFeedModel()
	m.Focused = true
	events := []bus.EventSummary{
		{Type: "graph.generated"},
		{Type: "trust.completed"},
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = next
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("graph")})
	m = next
	require.True(t, m.SearchMode)
	require.Equal(t, "graph", m.Search)

	selected := m.SelectedEvent(events)
	require.NotNil(t, selected)
	require.Equal(t, "graph.generated", selected.Type)
}

func TestEventFeedModelCyclesFilterAndPause(t *testing.T) {
	m := NewEventFeedModel()
	m.Focused = true

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = next
	require.Equal(t, "runtime", bus.EventFeedFilterTypes[m.FilterIndex])

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next
	require.True(t, m.Paused)
}

func TestRenderEventFeedShowsLiveControls(t *testing.T) {
	now := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	got := RenderEventFeed(EventFeedViewModel{
		Events: []bus.EventSummary{
			{Type: "graph.generated", CreatedAt: now},
			{Type: "trust.completed", CreatedAt: now.Add(time.Minute)},
		},
		Limit:  2,
		Filter: "all",
		Search: "(none)",
	})
	require.Contains(t, got, "Filter: all")
	require.Contains(t, got, "[live]")
	require.NotContains(t, got, "(stub)")
	require.Contains(t, got, "graph.generated")
}

func TestArtifactNavigationMapsGraphEvents(t *testing.T) {
	screen, cli := ArtifactNavigation(bus.EventSummary{Type: "graph.generated"})
	require.Equal(t, "graph", screen)
	require.Equal(t, "asa graph", cli)
}
