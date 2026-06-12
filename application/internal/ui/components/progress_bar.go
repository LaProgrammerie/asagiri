package components

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
)

const defaultProgressWidth = 10

// ProgressBar renders a ratio meter (0..1).
func ProgressBar(ratio float64, width int) string {
	return ProgressBarThemed(ratio, width, theme.Default())
}

// ProgressBarThemed renders a themed progress bar.
func ProgressBarThemed(ratio float64, width int, th theme.Theme) string {
	if width <= 0 {
		width = defaultProgressWidth
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	st := th.Styles()
	barStyle := st.Success
	if ratio < 0.5 {
		barStyle = st.Error
	} else if ratio < 0.8 {
		barStyle = st.Warning
	}
	return barStyle.Render(strings.Repeat("█", filled)) + st.Divider.Render(strings.Repeat("░", width-filled))
}
