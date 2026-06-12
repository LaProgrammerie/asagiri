package app

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
)

var layoutCycle = []layout.Kind{
	layout.SplitVertical,
	layout.Grid,
	layout.Dashboard,
	layout.Focus,
	layout.Fullscreen,
	layout.Single,
}

func (m model) currentLayout() layout.Kind {
	if m.state.Layout != "" {
		return m.state.Layout
	}
	if m.router.Current() == ScreenSettings {
		return layout.Single
	}
	if m.width > m.layout.CompactThreshold {
		return layout.SplitVertical
	}
	return layout.SplitHorizontal
}

func (m model) computeLayout() layout.Computed {
	kind := m.currentLayout()
	opts := layout.ComputeOpts{
		Grid:      layout.GridSpec{Columns: 2, Rows: 2},
		Dashboard: layout.DashboardSpec{Columns: layout.DashboardColumns(m.width, m.layout.CompactThreshold)},
		FocusPane: m.state.Focus,
	}
	computed := m.layout.ComputeWithOpts(kind, maxInt(1, m.width), maxInt(1, m.height), opts)
	return layout.Computed{
		Kind:  computed.Kind,
		Panes: computed.VisiblePanes(m.state.Panels),
	}
}

func (m *model) cycleLayoutMode() {
	current := m.state.Layout
	if current == "" {
		current = m.currentLayout()
	}
	next := layoutCycle[0]
	for i, k := range layoutCycle {
		if k == current {
			next = layoutCycle[(i+1)%len(layoutCycle)]
			break
		}
	}
	if next == layout.SplitVertical || next == layout.SplitHorizontal {
		m.state.Layout = ""
		m.lastCommandResult = "layout: auto"
		return
	}
	m.state.Layout = next
	m.lastCommandResult = fmt.Sprintf("layout: %s", next)
}

func (m *model) toggleFullscreen() {
	if m.state.Layout == layout.Fullscreen {
		m.state.Layout = ""
		m.lastCommandResult = "layout: auto"
		return
	}
	m.state.Layout = layout.Fullscreen
	m.lastCommandResult = "layout: fullscreen"
}

func (m *model) togglePaneCollapse() {
	m.state.Panels.ToggleCollapse(m.state.Focus)
	state := "expanded"
	if m.state.Panels.IsCollapsed(m.state.Focus) {
		state = "collapsed"
	}
	m.lastCommandResult = fmt.Sprintf("pane %s: %s", m.state.Focus, state)
}
