package state

import "github.com/LaProgrammerie/asagiri/application/internal/ui/layout"

// UIState stores local navigation state for the Bubble Tea shell.
type UIState struct {
	Screen string
	Layout layout.Kind
	Focus  layout.PaneID
}

// New creates a local UI state store.
func New(initialScreen string) UIState {
	if initialScreen == "" {
		initialScreen = "mission"
	}
	return UIState{
		Screen: initialScreen,
		Layout: layout.Single,
		Focus:  layout.PaneMain,
	}
}

// SetScreen updates the active screen id.
func (s *UIState) SetScreen(screen string) {
	if screen == "" {
		return
	}
	s.Screen = screen
}
