package state

import (
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/layout"
)

// UIState stores local navigation state for the Bubble Tea shell.
type UIState struct {
	Screen       string
	Layout       layout.Kind
	Focus        layout.PaneID
	Nav          NavigationStack
	Panels       layout.Session
	FocusContext bus.FocusContext
}

// NavigationFrame is one drill-down level inside an explorer screen.
type NavigationFrame struct {
	Screen  string
	Subject string
	Detail  string
}

// NavigationStack tracks explorer drill-down breadcrumbs.
type NavigationStack struct {
	Frames []NavigationFrame
}

// New creates a local UI state store.
func New(initialScreen string) UIState {
	if initialScreen == "" {
		initialScreen = "mission"
	}
	return UIState{
		Screen: initialScreen,
		Layout: "", // auto: app picks split by terminal width
		Focus:  layout.PaneMain,
		Panels: layout.NewSession(),
	}
}

// SetScreen updates the active screen id.
func (s *UIState) SetScreen(screen string) {
	if screen == "" {
		return
	}
	s.Screen = screen
}

// Push adds a drill-down frame.
func (s *NavigationStack) Push(frame NavigationFrame) {
	if frame.Screen == "" && frame.Subject == "" && frame.Detail == "" {
		return
	}
	s.Frames = append(s.Frames, frame)
}

// Pop removes the latest drill-down frame.
func (s *NavigationStack) Pop() (NavigationFrame, bool) {
	if len(s.Frames) == 0 {
		return NavigationFrame{}, false
	}
	idx := len(s.Frames) - 1
	frame := s.Frames[idx]
	s.Frames = s.Frames[:idx]
	return frame, true
}

// Top returns the latest frame, if any.
func (s NavigationStack) Top() (NavigationFrame, bool) {
	if len(s.Frames) == 0 {
		return NavigationFrame{}, false
	}
	return s.Frames[len(s.Frames)-1], true
}

// Reset clears drill-down history.
func (s *NavigationStack) Reset() {
	s.Frames = nil
}

// SetFocusContext stores the current explainability focus target.
func (s *UIState) SetFocusContext(ctx bus.FocusContext) {
	s.FocusContext = ctx
}
