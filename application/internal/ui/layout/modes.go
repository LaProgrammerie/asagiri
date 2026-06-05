package layout

// Extended layout kinds (spec-ui §9).
const (
	Grid       Kind = "grid"
	Dashboard  Kind = "dashboard"
	Focus      Kind = "focus"
	Fullscreen Kind = "fullscreen"
)

// GridSpec configures a fixed column grid.
type GridSpec struct {
	Columns int
	Rows    int
}

// DashboardSpec configures dashboard widget grid placement.
type DashboardSpec struct {
	Columns int // 1, 2, or 3
}

func gridLayout(width, height int, spec GridSpec) Computed {
	cols := spec.Columns
	if cols <= 0 {
		cols = 2
	}
	rows := spec.Rows
	if rows <= 0 {
		rows = 2
	}
	cellW := width / cols
	if cellW < 1 {
		cellW = 1
	}
	cellH := height / rows
	if cellH < 1 {
		cellH = 1
	}
	panes := make([]PaneBounds, 0, cols*rows)
	idx := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			id := PaneID(string(rune('a' + idx)))
			switch idx {
			case 0:
				id = PaneMain
			case 1:
				id = PaneSide
			}
			panes = append(panes, PaneBounds{
				ID:     id,
				X:      c * cellW,
				Y:      r * cellH,
				Width:  cellW,
				Height: cellH,
			})
			idx++
		}
	}
	return Computed{Kind: Grid, Panes: panes}
}

func dashboardLayout(width, height int, spec DashboardSpec) Computed {
	cols := spec.Columns
	if cols <= 0 {
		cols = 2
	}
	if cols > 3 {
		cols = 3
	}
	rows := 3
	if cols == 1 {
		rows = 6
	}
	cellW := width / cols
	if cellW < 1 {
		cellW = 1
	}
	cellH := height / rows
	if cellH < 1 {
		cellH = 1
	}
	panes := make([]PaneBounds, 0, cols*rows)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			id := PaneID(string(rune('a' + r*cols + c)))
			if r == 0 && c == 0 {
				id = PaneMain
			}
			panes = append(panes, PaneBounds{
				ID:     id,
				X:      c * cellW,
				Y:      r * cellH,
				Width:  cellW,
				Height: cellH,
			})
		}
	}
	return Computed{Kind: Dashboard, Panes: panes}
}

func focusLayout(width, height int, focus PaneID) Computed {
	sideW := width / 4
	if sideW < 8 {
		sideW = 8
	}
	mainW := width - sideW
	if mainW < 1 {
		mainW = 1
	}
	return Computed{
		Kind: Focus,
		Panes: []PaneBounds{
			{ID: focus, X: 0, Y: 0, Width: mainW, Height: height},
			{ID: PaneSide, X: mainW, Y: 0, Width: sideW, Height: height},
		},
	}
}

func fullscreenLayout(width, height int) Computed {
	return Computed{
		Kind: Fullscreen,
		Panes: []PaneBounds{
			{ID: PaneMain, X: 0, Y: 0, Width: width, Height: height},
		},
	}
}
