package layout

// Kind is one supported lot-1 layout mode.
type Kind string

const (
	Single          Kind = "single"
	SplitHorizontal Kind = "split-horizontal"
	SplitVertical   Kind = "split-vertical"
)

// PaneID identifies one pane in a computed layout.
type PaneID string

const (
	PaneMain PaneID = "main"
	PaneSide PaneID = "side"
)

// PaneBounds stores pane geometry in terminal cells.
type PaneBounds struct {
	ID     PaneID
	X      int
	Y      int
	Width  int
	Height int
}

// Computed holds one layout computation result.
type Computed struct {
	Kind  Kind
	Panes []PaneBounds
}

// Engine computes pane layouts.
type Engine struct {
	CompactThreshold int
}

// NewEngine returns a layout engine with sane defaults.
func NewEngine(compactThreshold int) Engine {
	if compactThreshold <= 0 {
		compactThreshold = 100
	}
	return Engine{CompactThreshold: compactThreshold}
}

// ComputeOpts carries optional parameters for advanced layout modes.
type ComputeOpts struct {
	Grid      GridSpec
	Dashboard DashboardSpec
	FocusPane PaneID
}

// Compute resolves geometry for one requested kind and terminal size.
func (e Engine) Compute(kind Kind, width, height int) Computed {
	return e.ComputeWithOpts(kind, width, height, ComputeOpts{})
}

// ComputeWithOpts resolves geometry with mode-specific options.
func (e Engine) ComputeWithOpts(kind Kind, width, height int, opts ComputeOpts) Computed {
	if width <= 0 || height <= 0 {
		return Computed{Kind: Single}
	}
	if width <= e.CompactThreshold && kind == SplitVertical {
		kind = Single
	}
	switch kind {
	case SplitHorizontal:
		return splitHorizontal(width, height)
	case SplitVertical:
		return splitVertical(width, height)
	case Grid:
		return gridLayout(width, height, opts.Grid)
	case Dashboard:
		return dashboardLayout(width, height, opts.Dashboard)
	case Focus:
		focus := opts.FocusPane
		if focus == "" {
			focus = PaneMain
		}
		return focusLayout(width, height, focus)
	case Fullscreen:
		return fullscreenLayout(width, height)
	default:
		return single(width, height)
	}
}

// DashboardColumns picks column count from terminal width (spec-ui §27).
func DashboardColumns(width, compactThreshold int) int {
	if width <= 0 {
		return 1
	}
	if width < compactThreshold {
		return 1
	}
	if width >= 180 {
		return 3
	}
	if width >= 120 {
		return 2
	}
	return 1
}
