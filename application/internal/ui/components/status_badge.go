package components

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// StatusBadge renders a labeled status chip.
func StatusBadge(label, status string, animated bool, th theme.Theme) string {
	state := ParseVisualState(status)
	glyph := StatusGlyph(state, animated)
	text := fmt.Sprintf("%s %s", glyph, label)
	color := th.Palette.Muted
	switch state {
	case StateRunning:
		color = th.Palette.Primary
	case StateSuccess:
		color = th.Palette.Success
	case StateWarning:
		color = th.Palette.Warning
	case StateError, StateBlocked:
		color = th.Palette.Error
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(text)
}
