package components

import (
	"fmt"
	"strings"
)

var shimmerFrames = []string{"░", "▒", "▓", "▒"}

// ShimmerPrefix returns a minimal loading prefix when animations are enabled.
func ShimmerPrefix(animated bool, frame int) string {
	if !animated {
		return ""
	}
	if len(shimmerFrames) == 0 {
		return ""
	}
	return shimmerFrames[frame%len(shimmerFrames)] + " "
}

// LoadingShimmer returns ShimmerPrefix for running or waiting visual states.
func LoadingShimmer(animated bool, frame int, state VisualState) string {
	switch state {
	case StateRunning, StateWaiting:
		return ShimmerPrefix(animated, frame)
	default:
		return ""
	}
}

// Sparkline renders a compact ASCII sparkline from normalized samples (0..1).
func Sparkline(samples []float64, width int) string {
	if width <= 0 {
		width = 8
	}
	if len(samples) == 0 {
		return strings.Repeat("·", width)
	}
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	out := make([]rune, width)
	for i := 0; i < width; i++ {
		idx := i * len(samples) / width
		if idx >= len(samples) {
			idx = len(samples) - 1
		}
		v := samples[idx]
		if v < 0 {
			v = 0
		}
		if v > 1 {
			v = 1
		}
		out[i] = chars[int(v*float64(len(chars)-1))]
	}
	return string(out)
}

// LiveCounter formats a counter with optional delta hint.
func LiveCounter(label string, value int, delta int) string {
	if delta == 0 {
		return fmt.Sprintf("%s: %d", label, value)
	}
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s: %d (%s%d)", label, value, sign, delta)
}

// FocusMarker prefixes content when a pane has keyboard focus.
func FocusMarker(focused bool, body string) string {
	if !focused {
		return body
	}
	return "▸ " + body
}
