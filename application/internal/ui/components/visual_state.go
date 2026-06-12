package components

import "strings"

// VisualState is one terminal status glyph set (spec-ui §8.3).
type VisualState string

const (
	StateIdle    VisualState = "idle"
	StateRunning VisualState = "running"
	StateSuccess VisualState = "success"
	StateWarning VisualState = "warning"
	StateError   VisualState = "error"
	StateBlocked VisualState = "blocked"
	StatePaused  VisualState = "paused"
	StateWaiting VisualState = "waiting"
	StateUnknown VisualState = "unknown"
)

// ParseVisualState maps free-text status to a visual state.
func ParseVisualState(status string) VisualState {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running", "active", "in_progress", "in-progress":
		return StateRunning
	case "done", "completed", "success", "succeeded", "ok":
		return StateSuccess
	case "warn", "warning", "degraded":
		return StateWarning
	case "failed", "error", "critical":
		return StateError
	case "blocked":
		return StateBlocked
	case "paused":
		return StatePaused
	case "waiting", "pending", "queued":
		return StateWaiting
	case "idle", "stopped":
		return StateIdle
	default:
		return StateUnknown
	}
}

// StatusGlyph returns a compact glyph for the state (animated spinners when enabled).
func StatusGlyph(state VisualState, animated bool) string {
	switch state {
	case StateRunning:
		if animated {
			return "⠋"
		}
		return "◐"
	case StateSuccess:
		return "✓"
	case StateWarning:
		return "!"
	case StateError:
		return "✕"
	case StateBlocked:
		return "⊘"
	case StatePaused:
		return "‖"
	case StateWaiting:
		return "○"
	case StateIdle:
		return "·"
	default:
		return "?"
	}
}
