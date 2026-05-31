package app

import (
	"strconv"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
)

// navRailWidth is the fixed width of the persistent navigation rail (cells).
const navRailWidth = 24

// navDest is one persistent rail destination wired to the router.
type navDest struct {
	screen string
	label  string
	icon   string
}

// navDestinations lists the rail entries in display order. They mirror the
// real router screens so the rail stays in sync with navigation.
func navDestinations() []navDest {
	return []navDest{
		{ScreenMission, "Mission", "◆"},
		{ScreenDashboard, "Dashboard", "▣"},
		{ScreenRuns, "Runs", "▶"},
		{ScreenAgents, "Agents", "◎"},
		{ScreenFlow, "Flow", "⇄"},
		{ScreenGraph, "Graph", "⊹"},
		{ScreenTrust, "Trust", "⛉"},
		{ScreenKnowledge, "Knowledge", "❏"},
		{ScreenReplay, "Replay", "↺"},
		{ScreenLogs, "Logs", "▤"},
		{ScreenExplain, "Explain", "?"},
		{ScreenSettings, "Settings", "⚙"},
	}
}

// railWidth returns the rail footprint, or 0 when the rail is collapsed
// (compact terminals, wizard, or while a modal/palette is open).
func (m model) railWidth() int {
	if m.wizardMode || m.showHelp || m.showPalette || m.confirmation != nil {
		return 0
	}
	if m.width < m.layout.CompactThreshold {
		return 0
	}
	return navRailWidth
}

// bodyWidth is the width budget available to the current screen, accounting for
// the persistent rail when visible.
func (m model) bodyWidth() int {
	w := m.width - m.railWidth()
	if w < minPaneWidthCells {
		w = minPaneWidthCells
	}
	return w
}

// navRailItems builds the rail entries with the active highlight and
// snapshot-derived state badges.
func (m model) navRailItems() []components.NavItem {
	current := m.router.Current()
	items := make([]components.NavItem, 0, len(navDestinations()))
	for _, d := range navDestinations() {
		items = append(items, components.NavItem{
			Icon:   d.icon,
			Label:  d.label,
			Badge:  m.navBadge(d.screen),
			Active: d.screen == current,
		})
	}
	return items
}

// navBadge returns the state badge for a destination (CK-2.2): active runs,
// trust alert, queued events.
func (m model) navBadge(screen string) string {
	switch screen {
	case ScreenRuns:
		if n := activeRunsCount(m.snapshot.Runs); n > 0 {
			return strconv.Itoa(n)
		}
	case ScreenAgents:
		if n := len(m.snapshot.ActiveAgents); n > 0 {
			return strconv.Itoa(n)
		}
	case ScreenTrust:
		if m.snapshot.Trust.Overall > 0 && m.snapshot.Trust.Overall < 0.5 {
			return "!"
		}
	case ScreenLogs:
		if q := m.snapshot.Runtime.Status.QueuedEvents; q > 0 {
			return strconv.Itoa(q)
		}
	}
	return ""
}

func activeRunsCount(runs []bus.RunSummary) int {
	n := 0
	for _, r := range runs {
		if r.Status == "running" {
			n++
		}
	}
	return n
}

// renderNavRail renders the persistent rail clamped to w×h.
func (m model) renderNavRail(w, h int) string {
	return components.RenderNavRail("NAVIGATION", m.navRailItems(), w, h, m.theme)
}

// railScreenForRow maps a rail content row (0-based, relative to the first nav
// entry) to its destination screen, or "" when out of range. The rail prints a
// "NAVIGATION" head plus a blank line before the entries.
func railScreenForRow(row int) string {
	dests := navDestinations()
	if row < 0 || row >= len(dests) {
		return ""
	}
	return dests[row].screen
}

// railFirstItemOffset is the number of rail rows before the first nav entry
// (section head + blank line).
const railFirstItemOffset = 2

// railFirstItemY returns the absolute terminal row of the first rail entry,
// mirroring the View() frame composition (border, title, status bar, optional
// tab bar) so mouse clicks map to the correct destination.
func (m model) railFirstItemY() int {
	top := 1 // frame top border
	top += 2 // hero title (title + meta lines)
	top += 1 // status bar
	top += 1 // blank line before tab bar / body
	if m.renderTabBar() != "" {
		top += 2 // tab bar + trailing blank line
	}
	return top + railFirstItemOffset
}

// railScreenAt returns the destination screen for a click at (x, y) within the
// rail, or "" when the click is outside the rail entries.
func (m model) railScreenAt(x, y int) string {
	rw := m.railWidth()
	if rw <= 0 || x < 0 || x >= rw {
		return ""
	}
	return railScreenForRow(y - m.railFirstItemY())
}
