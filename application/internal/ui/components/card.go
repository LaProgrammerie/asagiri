package components

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// MetricCard renders one compact metric.
func MetricCard(label string, value any, th theme.Theme) string {
	l := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Palette.Muted)).
		Render(label)
	v := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(th.Palette.Primary)).
		Render(fmt.Sprintf("%v", value))
	return Panel("Metric", l+": "+v, th)
}
