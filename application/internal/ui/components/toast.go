package components

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// ToastLevel is toast severity.
type ToastLevel string

const (
	ToastInfo    ToastLevel = "info"
	ToastSuccess ToastLevel = "success"
	ToastWarning ToastLevel = "warning"
	ToastError   ToastLevel = "error"
)

// RenderToast renders a transient notification line.
func RenderToast(level ToastLevel, message string, th theme.Theme) string {
	color := th.Palette.Muted
	switch level {
	case ToastSuccess:
		color = th.Palette.Success
	case ToastWarning:
		color = th.Palette.Warning
	case ToastError:
		color = th.Palette.Error
	case ToastInfo:
		color = th.Palette.Primary
	}
	glyph := StatusGlyph(StateUnknown, false)
	switch level {
	case ToastSuccess:
		glyph = StatusGlyph(StateSuccess, false)
	case ToastWarning:
		glyph = StatusGlyph(StateWarning, false)
	case ToastError:
		glyph = StatusGlyph(StateError, false)
	}
	line := fmt.Sprintf("%s %s", glyph, message)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(line)
}
