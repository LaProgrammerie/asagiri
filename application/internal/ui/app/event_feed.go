package app

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) eventFeedScreenActive() bool {
	switch m.router.Current() {
	case ScreenMission, ScreenDashboard:
		return true
	default:
		return false
	}
}

func (m model) eventFeedViewModel(limit int) components.EventFeedViewModel {
	vm := m.eventFeed.ViewModel(m.snapshot.Events, limit, m.cfg.ShowCLIEquivalents)
	vm.Focused = m.eventFeedFocused
	return vm
}

func (m *model) updateEventFeedKey(v tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.eventFeedScreenActive() {
		return m, nil
	}
	key := v.String()
	feedKey := key == "e" || key == "x" || key == "o" || key == "f" || key == "p" || key == " " ||
		key == "/" || key == "up" || key == "down" || key == "esc" || key == "backspace" ||
		(m.eventFeedFocused && m.eventFeed.SearchMode && v.Type == tea.KeyRunes)
	if !feedKey && !m.eventFeedFocused {
		return m, nil
	}
	switch key {
	case "e":
		m.eventFeedFocused = !m.eventFeedFocused
		m.eventFeed.Focused = m.eventFeedFocused
		return m, nil
	case "x":
		if !m.eventFeedFocused {
			return m, nil
		}
		filter := "all"
		if m.eventFeed.FilterIndex >= 0 && m.eventFeed.FilterIndex < len(bus.EventFeedFilterTypes) {
			filter = bus.EventFeedFilterTypes[m.eventFeed.FilterIndex]
		}
		return m, m.dispatchCommand(bus.ExportEventsCommand{
			TypeFilter: filter,
			Search:     strings.TrimSpace(m.eventFeed.Search),
		}, "asa runtime events --export")
	case "o":
		if !m.eventFeedFocused {
			return m, nil
		}
		ev := m.eventFeed.SelectedEvent(m.snapshot.Events)
		if ev == nil {
			return m, nil
		}
		screen, cli := components.ArtifactNavigation(*ev)
		m.navigateTo(screen, cli)
		return m, nil
	}
	if !m.eventFeedFocused {
		return m, nil
	}
	next, cmd := m.eventFeed.Update(v)
	m.eventFeed = next
	return m, cmd
}
