package components

import (
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Panel renders a titled container.
func Panel(title, body string, th theme.Theme) string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(th.Palette.Primary)).
		Render(title)
	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color(th.Palette.Muted)).
		Render(body)

	return th.BorderStyle().
		Padding(0, 1).
		Render(header + "\n" + content)
}
