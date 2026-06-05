package components

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
)

// EventFeedViewModel configures reusable event feed rendering.
type EventFeedViewModel struct {
	Events       []bus.EventSummary
	Limit        int
	Filter       string
	Search       string
	Paused       bool
	Cursor       int
	Focused      bool
	ShowCLIHints bool
}

// EventFeedModel is an interactive event feed controller (spec-ui §22).
type EventFeedModel struct {
	FilterIndex int
	Search      string
	SearchMode  bool
	Paused      bool
	Cursor      int
	Focused     bool
}

// NewEventFeedModel returns default event feed interaction state.
func NewEventFeedModel() EventFeedModel {
	return EventFeedModel{}
}

// ViewModel projects model state for rendering.
func (m EventFeedModel) ViewModel(events []bus.EventSummary, limit int, showCLI bool) EventFeedViewModel {
	filter := "all"
	if m.FilterIndex >= 0 && m.FilterIndex < len(bus.EventFeedFilterTypes) {
		filter = bus.EventFeedFilterTypes[m.FilterIndex]
	}
	search := strings.TrimSpace(m.Search)
	if search == "" {
		search = "(none)"
	}
	return EventFeedViewModel{
		Events:       events,
		Limit:        limit,
		Filter:       filter,
		Search:       search,
		Paused:       m.Paused,
		Cursor:       m.Cursor,
		Focused:      m.Focused,
		ShowCLIHints: showCLI,
	}
}

// Update handles event feed key bindings when the feed is focused.
func (m EventFeedModel) Update(msg tea.Msg) (EventFeedModel, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !m.Focused {
		return m, nil
	}
	switch key.String() {
	case "f":
		m.FilterIndex = (m.FilterIndex + 1) % len(bus.EventFeedFilterTypes)
		m.Cursor = 0
	case "p", " ":
		m.Paused = !m.Paused
	case "/":
		m.SearchMode = true
	case "esc":
		m.SearchMode = false
		m.Search = ""
		m.Cursor = 0
	case "up":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down":
		m.Cursor++
	case "backspace":
		if m.SearchMode && len(m.Search) > 0 {
			m.Search = m.Search[:len(m.Search)-1]
		}
		m.Cursor = 0
	default:
		if m.SearchMode && key.Type == tea.KeyRunes && len(key.Runes) > 0 {
			m.Search += string(key.Runes)
			m.Cursor = 0
		}
	}
	return m, nil
}

// SelectRow maps a terminal row to a filtered event cursor.
func (m *EventFeedModel) SelectRow(row int) {
	if row < 0 {
		return
	}
	m.Cursor = row
}

// SelectedEvent returns the event at the cursor after filtering, if any.
func (m EventFeedModel) SelectedEvent(events []bus.EventSummary) *bus.EventSummary {
	filtered := filterEvents(events, m.ViewModel(events, len(events), false))
	if len(filtered) == 0 {
		return nil
	}
	idx := m.Cursor
	if idx < 0 {
		idx = 0
	}
	if idx >= len(filtered) {
		idx = len(filtered) - 1
	}
	ev := filtered[idx]
	return &ev
}

// ArtifactNavigation returns screen id and CLI for the selected event.
func ArtifactNavigation(ev bus.EventSummary) (screen string, cli string) {
	typ := strings.ToLower(ev.Type)
	switch {
	case strings.HasPrefix(typ, "graph."):
		return "graph", "asa graph"
	case strings.HasPrefix(typ, "trust."):
		return "trust", "asa trust"
	case strings.HasPrefix(typ, "agent."):
		return "agents", "asa agents watch"
	case strings.HasPrefix(typ, "investigation."):
		return "explain", `asa investigate "event"`
	case strings.HasPrefix(typ, "replay."):
		return "replay", "asa replay open <replay-id>"
	case strings.HasPrefix(typ, "knowledge."):
		return "knowledge", "asa knowledge"
	case strings.HasPrefix(typ, "prototype."):
		return "prototype", "asa prototype"
	default:
		return "logs", "asa logs"
	}
}

// RenderEventFeed renders a reusable event feed block.
func RenderEventFeed(vm EventFeedViewModel) string {
	limit := vm.Limit
	if limit <= 0 {
		limit = 5
	}
	filter := strings.TrimSpace(vm.Filter)
	if filter == "" {
		filter = "all"
	}
	search := strings.TrimSpace(vm.Search)
	if search == "" {
		search = "(none)"
	}

	var b strings.Builder
	mode := "live"
	if vm.Paused {
		mode = "paused"
	}
	focus := ""
	if vm.Focused {
		focus = " [focused]"
	}
	fmt.Fprintf(&b, "Filter: %s  Search: %s  [%s]%s\n", filter, search, mode, focus)
	if vm.ShowCLIHints {
		b.WriteString("Keys: f filter  / search  p pause  x export  o open  e focus  ↑↓ select\n")
		b.WriteString("CLI: asa runtime events --type <filter> --search <query> --export\n")
	}

	events := vm.Events
	if !vm.Paused {
		events = filterEvents(events, vm)
	}
	if len(events) == 0 {
		b.WriteString("- none")
		return b.String()
	}

	filtered := filterEvents(events, vm)
	cursor := vm.Cursor
	if cursor < 0 {
		cursor = 0
	}
	win := SliceWindow(len(filtered), cursor, limit)
	visible := VisibleSlice(filtered, win)
	for i, ev := range visible {
		idx := win.Offset + i
		prefix := " "
		if idx == cursor && vm.Focused {
			prefix = ">"
		}
		fmt.Fprintf(&b, "%s %s  %s\n", prefix, ev.CreatedAt.Format("15:04:05"), ev.Type)
	}
	if win.Total > win.Limit {
		fmt.Fprintf(&b, "… %d more (virtual)\n", win.Total-win.Limit)
	}
	count := len(visible)
	if count == 0 {
		b.WriteString("- none")
	}
	return strings.TrimRight(b.String(), "\n")
}

func filterEvents(events []bus.EventSummary, vm EventFeedViewModel) []bus.EventSummary {
	out := make([]bus.EventSummary, 0, len(events))
	for _, ev := range events {
		if !eventMatchesFilter(ev, vm.Filter) || !eventMatchesSearch(ev, vm.Search) {
			continue
		}
		out = append(out, ev)
	}
	return out
}

func eventMatchesFilter(ev bus.EventSummary, filter string) bool {
	f := strings.ToLower(strings.TrimSpace(filter))
	if f == "" || f == "all" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(ev.Type), f+".") || strings.Contains(strings.ToLower(ev.Type), f)
}

func eventMatchesSearch(ev bus.EventSummary, search string) bool {
	s := strings.ToLower(strings.TrimSpace(search))
	if s == "" || s == "(none)" {
		return true
	}
	return strings.Contains(strings.ToLower(ev.Type), s) ||
		strings.Contains(strings.ToLower(ev.Source), s) ||
		strings.Contains(strings.ToLower(ev.SessionID), s) ||
		strings.Contains(strings.ToLower(ev.FlowID), s)
}
