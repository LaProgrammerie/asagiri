package components

import "strings"

const defaultProgressWidth = 10

// ProgressBar renders a ratio meter (0..1).
func ProgressBar(ratio float64, width int) string {
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
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
