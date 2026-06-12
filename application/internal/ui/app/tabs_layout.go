package app

import (
	"fmt"
	"slices"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/mission"
	"github.com/charmbracelet/lipgloss"
)

func (m model) screenTabLabels() []string {
	switch m.router.Current() {
	case ScreenMission:
		return []string{"Overview", "Agents", "Events"}
	case ScreenDashboard:
		if m.width >= m.layout.CompactThreshold {
			return []string{"Widgets", "Metrics"}
		}
	}
	return nil
}

func (m *model) syncScreenTabs() {
	labels := m.screenTabLabels()
	if slices.Equal(m.state.Panels.TabLabels, labels) {
		return
	}
	m.state.Panels.TabLabels = labels
	m.state.Panels.SetTab(0)
}

func (m model) activeScreenTab() int {
	labels := m.screenTabLabels()
	if len(labels) == 0 {
		return 0
	}
	if len(m.state.Panels.TabLabels) == 0 {
		return 0
	}
	tab := m.state.Panels.ActiveTab
	if tab < 0 {
		return 0
	}
	if tab >= len(labels) {
		return len(labels) - 1
	}
	return tab
}

func (m model) renderTabBar() string {
	labels := m.screenTabLabels()
	if len(labels) == 0 {
		return ""
	}
	return components.RenderTabs(components.TabsViewModel{
		Labels:  labels,
		Active:  m.activeScreenTab(),
		Focused: true,
		Theme:   m.theme,
	})
}

func (m *model) nextScreenTab() {
	m.syncScreenTabs()
	if len(m.state.Panels.TabLabels) == 0 {
		return
	}
	m.state.Panels.NextTab()
	m.lastCommandResult = fmt.Sprintf("tab: %s", m.state.Panels.TabLabels[m.state.Panels.ActiveTab])
}

func (m *model) prevScreenTab() {
	m.syncScreenTabs()
	if len(m.state.Panels.TabLabels) == 0 {
		return
	}
	m.state.Panels.SetTab(m.state.Panels.ActiveTab - 1)
	m.lastCommandResult = fmt.Sprintf("tab: %s", m.state.Panels.TabLabels[m.state.Panels.ActiveTab])
}

func (m model) renderMissionWithTabs(vm mission.ViewModel) string {
	tabBar := m.renderTabBar()
	body := m.renderMissionBody(vm, m.activeScreenTab())
	if tabBar == "" {
		return body
	}
	return strings.Join([]string{tabBar, body}, "\n")
}

func (m model) renderMissionBody(vm mission.ViewModel, tab int) string {
	headerPanel := components.Panel("ASAGIRI", mission.RenderHeader(vm), m.theme)
	runtimePanel := components.Panel("Runtime", mission.RenderRuntimeRuns(vm), m.theme)
	trustPanel := components.Panel("Trust", mission.RenderTrustPane(vm), m.theme)
	flowPanel := components.Panel("Active Flow", mission.RenderActiveFlowPane(vm), m.theme)
	agentsPanel := components.Panel("Agent Theatre", mission.RenderAgentTheatrePane(vm), m.theme)
	eventsPanel := components.Panel("Events", mission.RenderEventsPane(vm), m.theme)
	actionsPanel := components.Panel("Recommended actions", mission.RenderRecommendedActionsPane(vm), m.theme)

	switch tab {
	case 1:
		panes := []string{headerPanel, flowPanel, agentsPanel}
		return strings.Join(panes, "\n")
	case 2:
		panes := []string{headerPanel, actionsPanel, eventsPanel}
		return strings.Join(panes, "\n")
	default:
		// Rich panelised cockpit (CK-1): layout.Engine + dashboard widgets.
		// Plain/json (ui.mode) and zero geometry fall back to flat Render.
		if m.useRichLayout() && vm.Width > 0 && vm.Height > 0 {
			return mission.RenderCockpit(vm)
		}
		if !m.useRichLayout() {
			return mission.Render(vm)
		}
		return m.renderMissionOverview(vm, headerPanel, runtimePanel, trustPanel, flowPanel, agentsPanel, actionsPanel, eventsPanel)
	}
}

func (m model) renderMissionOverview(
	vm mission.ViewModel,
	headerPanel, runtimePanel, trustPanel, flowPanel, agentsPanel, actionsPanel, eventsPanel string,
) string {
	panes := []string{headerPanel}
	if m.width >= m.layout.CompactThreshold {
		if almostEqual(m.verticalSplit, defaultSplitRatio) {
			panes = append(panes, lipgloss.JoinHorizontal(lipgloss.Top, runtimePanel, trustPanel))
		} else {
			leftWidth, rightWidth := m.verticalPaneWidths()
			left := lipgloss.NewStyle().Width(leftWidth).Render(runtimePanel)
			right := lipgloss.NewStyle().Width(rightWidth).Render(trustPanel)
			panes = append(panes, lipgloss.JoinHorizontal(lipgloss.Top, left, right))
		}
	} else {
		panes = append(panes, runtimePanel, trustPanel)
	}
	panes = append(panes, flowPanel, agentsPanel, actionsPanel, eventsPanel)
	return strings.Join(panes, "\n")
}
