package layout

// Primary returns the main pane when available.
func (c Computed) Primary() (PaneBounds, bool) {
	for _, p := range c.Panes {
		if p.ID == PaneMain {
			return p, true
		}
	}
	if len(c.Panes) == 0 {
		return PaneBounds{}, false
	}
	return c.Panes[0], true
}

// Secondary returns the side pane when available.
func (c Computed) Secondary() (PaneBounds, bool) {
	for _, p := range c.Panes {
		if p.ID == PaneSide {
			return p, true
		}
	}
	return PaneBounds{}, false
}
