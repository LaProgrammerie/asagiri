package layout

// NextFocus cycles between panes for tab navigation.
func NextFocus(c Computed, current PaneID) PaneID {
	if len(c.Panes) == 0 {
		return PaneMain
	}
	idx := 0
	for i := range c.Panes {
		if c.Panes[i].ID == current {
			idx = i
			break
		}
	}
	return c.Panes[(idx+1)%len(c.Panes)].ID
}

// PrevFocus cycles backwards between panes for shift+tab navigation.
func PrevFocus(c Computed, current PaneID) PaneID {
	if len(c.Panes) == 0 {
		return PaneMain
	}
	idx := 0
	for i := range c.Panes {
		if c.Panes[i].ID == current {
			idx = i
			break
		}
	}
	prev := idx - 1
	if prev < 0 {
		prev = len(c.Panes) - 1
	}
	return c.Panes[prev].ID
}
