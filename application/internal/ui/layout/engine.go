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

// Compute resolves geometry for one requested kind and terminal size.
func (e Engine) Compute(kind Kind, width, height int) Computed {
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
	default:
		return single(width, height)
	}
}
