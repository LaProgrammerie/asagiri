package mission

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
	"github.com/charmbracelet/lipgloss"
)

const (
	cockpitGap         = 1
	cockpitMinPaneW    = 30
	cockpitMaxPaneRows = 12
	cockpitFrameMargin = 6
)

type cockpitPane struct {
	title string
	body  string
}

// RenderCockpit renders Mission Control as a responsive panelised grid built
// from the existing flat pane content and laid out via layout.DashboardColumns.
// It is the rich path; callers keep Render (flat) for plain/json parity.
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

	panes := []cockpitPane{
		{"Runtime", renderCockpitRuntime(vm)},
		{"Trust", dropHeading(RenderTrustPane(vm))},
		{"Agents", dropHeading(RenderAgentTheatrePane(vm))},
		{"Active Flow", dropHeading(RenderActiveFlowPane(vm))},
		{"Runs", renderRunsSummaryPane(vm)},
		{"Events", dropHeading(RenderEventsPane(vm))},
	}

	width := vm.Width - cockpitFrameMargin
	if width < cockpitMinPaneW {
		width = cockpitMinPaneW
	}
	colW := (width - (cols-1)*cockpitGap) / cols
	if colW < cockpitMinPaneW {
		colW = cockpitMinPaneW
	}

	header := renderCockpitHeader(vm, width)

	var rows []string
	for start := 0; start < len(panes); start += cols {
		end := start + cols
		if end > len(panes) {
			end = len(panes)
		}
		rows = append(rows, renderCockpitRow(vm, panes[start:end], colW))
	}
	grid := strings.Join(rows, "\n")
	if header == "" {
		return grid
	}
	return header + "\n" + grid
}

func renderCockpitRow(vm ViewModel, panes []cockpitPane, colW int) string {
	rowH := 0
	for _, p := range panes {
		if lines := lipgloss.Height(p.body) + 1; lines > rowH {
			rowH = lines
		}
	}
	rowH += 2 // panel border rows
	if rowH > cockpitMaxPaneRows {
		rowH = cockpitMaxPaneRows
	}
	if rowH < 4 {
		rowH = 4
	}

	rendered := make([]string, 0, len(panes)*2)
	for i, p := range panes {
		if i > 0 {
			rendered = append(rendered, strings.Repeat(" ", cockpitGap))
		}
		rendered = append(rendered, components.PanelSized(p.title, p.body, colW, rowH, vm.Theme))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
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

// renderRunsSummaryPane lists recent runs for the cockpit Runs pane.
func renderRunsSummaryPane(vm ViewModel) string {
	if len(vm.Runs) == 0 {
		return "No recent runs"
	}
	var b strings.Builder
	for i, run := range vm.Runs {
		if i >= 5 {
			break
		}
		feature := run.Feature
		if feature == "" {
			feature = run.ID
		}
		b.WriteString(fmt.Sprintf("%s  %s  %s\n", runStatusGlyph(run.Status), feature, run.Status))
	}
	return strings.TrimRight(b.String(), "\n")
}

// renderCockpitRuntime builds the runtime pane body without the recent-runs
// list (Runs has its own pane) and without the redundant heading.
func renderCockpitRuntime(vm ViewModel) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Agents: %d\n", len(vm.ActiveAgents)))
	sessions := 0
	if vm.SessionStatus == "active" {
		sessions = 1
	}
	b.WriteString(fmt.Sprintf("Sessions: %d\n", sessions))
	b.WriteString(fmt.Sprintf("Queue: %d\n", vm.QueuedEvents))
	b.WriteString(fmt.Sprintf("Cost today: €%.2f\n", vm.CostTodayEUR))
	b.WriteString(fmt.Sprintf("Cost month: €%.2f", vm.CostMonthEUR))
	if vm.Warning != "" || len(vm.Warnings) > 0 {
		b.WriteString("\nWarnings:")
		if vm.Warning != "" {
			b.WriteString("\n- " + vm.Warning)
		}
		for _, w := range vm.Warnings {
			b.WriteString("\n- " + w)
		}
	}
	return b.String()
}

// dropHeading removes the first line of a flat pane body so the panel title is
// not duplicated inside the cockpit panel.
func dropHeading(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return ""
}

func runStatusGlyph(status string) string {
	switch status {
	case "completed", "done", "success":
		return "✓"
	case "running":
		return "•"
	case "failed", "error":
		return "✕"
	case "blocked":
		return "⊘"
	default:
		return "○"
	}
}
