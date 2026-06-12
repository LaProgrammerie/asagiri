package dashboard

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ViewModel carries dashboard rendering inputs.
type ViewModel struct {
	Snapshot  bus.MissionControlSnapshotResult
	EventFeed components.EventFeedViewModel
	Theme     theme.Theme
	Width     int
	Height    int
	Animated  bool
	AnimFrame int
	TabIndex  int
	TabLabels []string
	Compact   int // layout compact threshold; 0 = default 100
}

// Render builds the live dashboard with composable widgets (spec-ui §12, §23).
func Render(vm ViewModel) string {
	var b strings.Builder
	if banner := components.ReadinessBanner(components.ReadinessBannerViewModel{
		Ready: vm.Snapshot.Readiness.Ready,
		Score: vm.Snapshot.Readiness.Score,
		Theme: vm.Theme,
	}); banner != "" {
		b.WriteString(banner)
		b.WriteString("\n\n")
	}
	if bar := renderDashboardTabBar(vm); bar != "" {
		b.WriteString(bar)
		b.WriteString("\n\n")
	}
	widgets := buildWidgets(vm)
	panels := widgetPanels(widgets, vm.Theme)
	threshold := vm.Compact
	if threshold <= 0 {
		threshold = 100
	}
	cols := layout.DashboardColumns(vm.Width, threshold)
	if cols <= 1 || vm.Width < threshold {
		b.WriteString(strings.Join(panels, "\n"))
		return b.String()
	}
	b.WriteString(renderDashboardGrid(panels, cols))
	return b.String()
}

func renderDashboardTabBar(vm ViewModel) string {
	if vm.Width < vm.compactThreshold() || len(vm.TabLabels) == 0 {
		return ""
	}
	return components.RenderTabs(components.TabsViewModel{
		Labels:  vm.TabLabels,
		Active:  vm.TabIndex,
		Focused: true,
		Theme:   vm.Theme,
	})
}

func (vm ViewModel) compactThreshold() int {
	if vm.Compact > 0 {
		return vm.Compact
	}
	return 100
}

func buildWidgets(vm ViewModel) []Widget {
	compact := vm.Width > 0 && vm.Width < 100
	if compact {
		return filterDashboardWidgets([]Widget{
			RuntimeWidget(vm.Snapshot, vm.Animated, vm.AnimFrame),
			TrustWidget(vm.Snapshot, vm.Animated),
			FlowWidget(vm.Snapshot, vm.Animated, vm.AnimFrame),
			RiskWidget(vm.Snapshot, vm.Animated),
			KnowledgeWidget(vm.Snapshot, vm.Animated),
			EventWidget(vm.Snapshot, vm.EventFeed, vm.Animated),
		}, vm.TabIndex)
	}
	base := []Widget{
		RuntimeWidget(vm.Snapshot, vm.Animated, vm.AnimFrame),
		SessionsWidget(vm.Snapshot, vm.Animated),
		QueueWidget(vm.Snapshot, vm.Animated),
		AgentWidget(vm.Snapshot, vm.Animated, vm.AnimFrame),
		TrustWidget(vm.Snapshot, vm.Animated),
		CostWidget(vm.Snapshot, vm.Animated),
		FlowWidget(vm.Snapshot, vm.Animated, vm.AnimFrame),
		RiskWidget(vm.Snapshot, vm.Animated),
		KnowledgeWidget(vm.Snapshot, vm.Animated),
		ReplayWidget(vm.Snapshot, vm.Animated),
		EventWidget(vm.Snapshot, vm.EventFeed, vm.Animated),
		ProgressWidget(vm.Snapshot, vm.Animated),
		PerformanceWidget(vm.Snapshot, vm.Animated),
	}
	return filterDashboardWidgets(base, vm.TabIndex)
}

func filterDashboardWidgets(widgets []Widget, tabIndex int) []Widget {
	if tabIndex != 1 {
		return widgets
	}
	metrics := map[string]bool{
		"Runtime": true, "Sessions": true, "Queue": true,
		"Progress": true, "Performance": true, "Costs": true,
	}
	out := make([]Widget, 0, len(widgets))
	for _, w := range widgets {
		if metrics[w.Title()] {
			out = append(out, w)
		}
	}
	if len(out) == 0 {
		return widgets
	}
	return out
}

func widgetPanels(widgets []Widget, th theme.Theme) []string {
	panels := make([]string, 0, len(widgets))
	for _, widget := range widgets {
		panels = append(panels, components.Panel(widget.Title(), widget.View(), th))
	}
	return panels
}

func renderDashboardGrid(panels []string, cols int) string {
	if cols < 2 {
		return strings.Join(panels, "\n")
	}
	rows := make([]string, 0, (len(panels)+cols-1)/cols)
	for i := 0; i < len(panels); i += cols {
		row := panels[i:]
		if len(row) > cols {
			row = row[:cols]
		}
		if len(row) == 1 {
			rows = append(rows, row[0])
			continue
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return strings.Join(rows, "\n")
}
