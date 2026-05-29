package dashboard

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ViewModel carries dashboard rendering inputs.
type ViewModel struct {
	Snapshot bus.MissionControlSnapshotResult
	Theme    theme.Theme
	Width    int
	Animated bool
}

// Render builds the lot-2 dashboard with V1 widgets.
func Render(vm ViewModel) string {
	widgets := buildWidgets(vm)
	panels := make([]string, 0, len(widgets))
	for _, widget := range widgets {
		panels = append(panels, components.Panel(widget.Title(), widget.View(), vm.Theme))
	}
	if vm.Width >= 180 {
		return renderUltraWide(panels)
	}
	if vm.Width >= 120 {
		return renderWide(panels)
	}
	return strings.Join(panels, "\n")
}

func buildWidgets(vm ViewModel) []Widget {
	compact := vm.Width > 0 && vm.Width < 100
	if compact {
		return []Widget{
			RuntimeWidget(vm.Snapshot, vm.Animated),
			TrustWidget(vm.Snapshot, vm.Animated),
			FlowWidget(vm.Snapshot, vm.Animated),
			EventWidget(vm.Snapshot, vm.Animated),
		}
	}
	return []Widget{
		RuntimeWidget(vm.Snapshot, vm.Animated),
		AgentWidget(vm.Snapshot, vm.Animated),
		TrustWidget(vm.Snapshot, vm.Animated),
		CostWidget(vm.Snapshot, vm.Animated),
		FlowWidget(vm.Snapshot, vm.Animated),
		EventWidget(vm.Snapshot, vm.Animated),
		ProgressWidget(vm.Snapshot, vm.Animated),
	}
}

func renderWide(panels []string) string {
	rows := make([]string, 0, (len(panels)+1)/2)
	for i := 0; i < len(panels); i += 2 {
		left := panels[i]
		right := ""
		if i+1 < len(panels) {
			right = panels[i+1]
		}
		if right == "" {
			rows = append(rows, left)
			continue
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, left, right))
	}
	return strings.Join(rows, "\n")
}

func renderUltraWide(panels []string) string {
	rows := make([]string, 0, (len(panels)+2)/3)
	for i := 0; i < len(panels); i += 3 {
		row := []string{panels[i]}
		if i+1 < len(panels) {
			row = append(row, panels[i+1])
		}
		if i+2 < len(panels) {
			row = append(row, panels[i+2])
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	return strings.Join(rows, "\n")
}
