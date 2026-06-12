package layout

// PanelPrefs stores per-pane UI preferences (session persistence).
type PanelPrefs struct {
	Collapsed bool
	TabIndex  int
	SizeRatio float64
}

// Session tracks layout session state for collapse/tabs/resize ratios.
type Session struct {
	Prefs     map[PaneID]PanelPrefs
	ActiveTab int
	TabLabels []string
}

// NewSession returns an empty panel session.
func NewSession() Session {
	return Session{Prefs: make(map[PaneID]PanelPrefs)}
}

// ToggleCollapse flips collapsed state for a pane.
func (s *Session) ToggleCollapse(id PaneID) {
	p := s.Prefs[id]
	p.Collapsed = !p.Collapsed
	s.Prefs[id] = p
}

// IsCollapsed reports whether a pane is collapsed.
func (s Session) IsCollapsed(id PaneID) bool {
	return s.Prefs[id].Collapsed
}

// SetTab switches the active tab index.
func (s *Session) SetTab(index int) {
	if len(s.TabLabels) == 0 {
		s.ActiveTab = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(s.TabLabels) {
		index = len(s.TabLabels) - 1
	}
	s.ActiveTab = index
}

// NextTab advances the active tab.
func (s *Session) NextTab() {
	if len(s.TabLabels) == 0 {
		return
	}
	s.ActiveTab = (s.ActiveTab + 1) % len(s.TabLabels)
}

// SizeRatio returns stored split ratio or defaultRatio.
func (s Session) SizeRatio(id PaneID, defaultRatio float64) float64 {
	if r := s.Prefs[id].SizeRatio; r > 0 {
		return r
	}
	return defaultRatio
}

// SetSizeRatio stores a resize ratio for a pane.
func (s *Session) SetSizeRatio(id PaneID, ratio float64) {
	p := s.Prefs[id]
	p.SizeRatio = ratio
	s.Prefs[id] = p
}

// VisiblePanes filters collapsed panes from computed layout.
func (c Computed) VisiblePanes(session Session) []PaneBounds {
	if len(c.Panes) == 0 {
		return nil
	}
	out := make([]PaneBounds, 0, len(c.Panes))
	for _, p := range c.Panes {
		if session.IsCollapsed(p.ID) {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return c.Panes[:1]
	}
	return out
}
