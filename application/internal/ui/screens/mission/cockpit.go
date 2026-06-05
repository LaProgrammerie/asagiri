package mission

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/screens/dashboard"
	"github.com/charmbracelet/lipgloss"
)

const (
	cockpitGap         = 1
	cockpitMinPaneW    = 30
	cockpitMaxPaneRows = 12
	cockpitFrameMargin = 6
)

// RenderCockpit renders Mission Control as a responsive panelised grid built
// from dashboard widgets and laid out via layout.Engine (CK-1.2).
func RenderCockpit(vm ViewModel) string {
	if vm.Width <= 0 || vm.Height <= 0 {
		return Render(vm)
	}
	compact := vm.CompactThreshold
	if compact <= 0 {
		compact = 100
	}
	cols := layout.DashboardColumns(vm.Width, compact)
	if cols < 1 {
		cols = 1
	}

	width := vm.Width - cockpitFrameMargin
	if width < cockpitMinPaneW {
		width = cockpitMinPaneW
	}
	header := renderCockpitHeader(vm, width)
	headerH := 0
	if header != "" {
		headerH = lipgloss.Height(header) + 1
	}
	contentH := vm.Height - headerH
	if contentH < 8 {
		contentH = 8
	}

	snap := snapshotFromVM(vm)
	animated := !vm.DisableAnimations
	widgets := []dashboard.Widget{
		dashboard.RuntimeWidget(snap, animated, vm.AnimFrame),
		dashboard.TrustWidget(snap, animated),
		dashboard.AgentWidget(snap, animated, vm.AnimFrame),
		dashboard.FlowWidget(snap, animated, vm.AnimFrame),
		dashboard.RunsSummaryWidget(snap),
		dashboard.EventWidget(snap, vm.EventFeed, animated),
	}

	engine := layout.NewEngine(compact)
	computed := engine.ComputeWithOpts(layout.Dashboard, width, contentH, layout.ComputeOpts{
		Dashboard: layout.DashboardSpec{Columns: cols},
	})

	panels := make([]string, 0, len(widgets))
	for i, w := range widgets {
		if i >= len(computed.Panes) {
			break
		}
		b := computed.Panes[i]
		paneW := b.Width - 2
		paneH := b.Height - 2
		if paneW < cockpitMinPaneW {
			paneW = cockpitMinPaneW
		}
		if paneH > cockpitMaxPaneRows {
			paneH = cockpitMaxPaneRows
		}
		if paneH < 4 {
			paneH = 4
		}
		title := w.Title()
		if title == "Flow" {
			title = "Active Flow"
		}
		panels = append(panels, components.PanelSized(title, w.View(), paneW, paneH, vm.Theme))
	}

	grid := renderCockpitGrid(panels, cols)

	// Onboarding invite (R7.6): when the repository is not onboarded, surface the
	// explicit `asa onboard` call to action. Presentation only — the readiness
	// state is sourced from the bus (ADR-027).
	banner := components.ReadinessBanner(components.ReadinessBannerFromResult(vm.Readiness))

	parts := make([]string, 0, 3)
	if header != "" {
		parts = append(parts, header)
	}
	if banner != "" {
		parts = append(parts, banner)
	}
	parts = append(parts, grid)
	return strings.Join(parts, "\n")
}

func renderCockpitGrid(panels []string, cols int) string {
	if cols < 1 {
		cols = 1
	}
	var rows []string
	for start := 0; start < len(panels); start += cols {
		end := start + cols
		if end > len(panels) {
			end = len(panels)
		}
		row := panels[start:end]
		if len(row) == 1 {
			rows = append(rows, row[0])
			continue
		}
		rendered := make([]string, 0, len(row)*2-1)
		for i, p := range row {
			if i > 0 {
				rendered = append(rendered, strings.Repeat(" ", cockpitGap))
			}
			rendered = append(rendered, p)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
	}
	return strings.Join(rows, "\n")
}

func snapshotFromVM(vm ViewModel) bus.MissionControlSnapshotResult {
	running := strings.HasPrefix(vm.RuntimeStatus, "running")
	sessions := 0
	if vm.SessionStatus == "active" {
		sessions = 1
	}
	return bus.MissionControlSnapshotResult{
		Workspace:     vm.Workspace,
		Branch:        vm.Branch,
		SessionStatus: vm.SessionStatus,
		Runtime: bus.RuntimeStatusResult{
			Status: runtime.DaemonStatus{
				Running:      running,
				Sessions:     sessions,
				QueuedEvents: vm.QueuedEvents,
			},
			Warning: vm.Warning,
		},
		Trust:        vm.Trust,
		Runs:         vm.Runs,
		Events:       vm.Events,
		ActiveAgents: vm.ActiveAgents,
		Flow:         vm.Flow,
		CostTodayEUR: vm.CostTodayEUR,
		CostMonthEUR: vm.CostMonthEUR,
		Readiness:    vm.Readiness,
	}
}

func renderCockpitHeader(vm ViewModel, width int) string {
	st := vm.Theme.Styles()
	workspace := vm.Workspace
	if workspace == "" {
		workspace = "-"
	}
	branch := vm.Branch
	if branch == "" {
		branch = "main"
	}
	left := st.Brand.Render(" ASAGIRI ") + " " +
		st.Muted.Render("⌂ "+workspace) + "  " + st.Muted.Render("⎇ "+branch)
	runtime := st.Success.Render(vm.RuntimeStatus)
	if !strings.HasPrefix(vm.RuntimeStatus, "running") {
		runtime = st.Muted.Render(vm.RuntimeStatus)
	}
	right := st.Muted.Render("runtime: ") + runtime +
		"  " + st.Muted.Render(fmt.Sprintf("€%.2f today", vm.CostTodayEUR))
	return components.RenderTopBar(left, right, width, vm.Theme)
}

// renderRunsSummaryPane lists recent runs (kept for unit tests).
func renderRunsSummaryPane(vm ViewModel) string {
	return dashboard.RunsSummaryWidget(snapshotFromVM(vm)).View()
}
