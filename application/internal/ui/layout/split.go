package layout

func single(width, height int) Computed {
	return Computed{
		Kind: Single,
		Panes: []PaneBounds{
			{ID: PaneMain, X: 0, Y: 0, Width: width, Height: height},
		},
	}
}

func splitHorizontal(width, height int) Computed {
	top := height / 2
	if top < 1 {
		top = 1
	}
	bottom := height - top
	if bottom < 1 {
		bottom = 1
		top = height - bottom
	}
	return Computed{
		Kind: SplitHorizontal,
		Panes: []PaneBounds{
			{ID: PaneMain, X: 0, Y: 0, Width: width, Height: top},
			{ID: PaneSide, X: 0, Y: top, Width: width, Height: bottom},
		},
	}
}

func splitVertical(width, height int) Computed {
	left := width / 2
	if left < 1 {
		left = 1
	}
	right := width - left
	if right < 1 {
		right = 1
		left = width - right
	}
	return Computed{
		Kind: SplitVertical,
		Panes: []PaneBounds{
			{ID: PaneMain, X: 0, Y: 0, Width: left, Height: height},
			{ID: PaneSide, X: left, Y: 0, Width: right, Height: height},
		},
	}
}
