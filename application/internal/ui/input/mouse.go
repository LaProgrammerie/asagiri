package input

import (
	"time"
)

const (
	doubleClickWindow = 400 * time.Millisecond
)

// MouseConfig controls mouse support.
type MouseConfig struct {
	Enabled bool
}

// DefaultMouseConfig returns default mouse behavior.
func DefaultMouseConfig() MouseConfig {
	return MouseConfig{Enabled: true}
}

// MenuItem is one context menu row.
type MenuItem struct {
	ID    string
	Label string
	CLI   string
}

// ContextMenu tracks an open contextual menu.
type ContextMenu struct {
	X, Y     int
	Items    []MenuItem
	Selected int
}

// SelectionRange tracks a list/text selection span.
type SelectionRange struct {
	Start int
	End   int
}

// MouseState tracks hover, click timing, menus, and selection.
type MouseState struct {
	LastClickAt time.Time
	LastClickX  int
	LastClickY  int
	HoverRow    int
	HoverCol    int
	ContextMenu *ContextMenu
	Selection   SelectionRange
}

// NewMouseState returns an empty mouse interaction state.
func NewMouseState() MouseState {
	return MouseState{HoverRow: -1, HoverCol: -1}
}

// IsDoubleClick reports whether two presses at nearby coordinates happened quickly.
func IsDoubleClick(prev time.Time, prevX, prevY, x, y int, now time.Time) bool {
	if prev.IsZero() {
		return false
	}
	if now.Sub(prev) > doubleClickWindow {
		return false
	}
	dx := prevX - x
	if dx < 0 {
		dx = -dx
	}
	dy := prevY - y
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1
}

// OpenContextMenu builds a menu at the given coordinates.
func OpenContextMenu(x, y int, items []MenuItem) *ContextMenu {
	if len(items) == 0 {
		return nil
	}
	copied := append([]MenuItem(nil), items...)
	return &ContextMenu{X: x, Y: y, Items: copied}
}

// MoveContextMenuSelection adjusts the highlighted menu row.
func MoveContextMenuSelection(menu *ContextMenu, delta int) {
	if menu == nil || len(menu.Items) == 0 {
		return
	}
	menu.Selected += delta
	if menu.Selected < 0 {
		menu.Selected = len(menu.Items) - 1
	}
	if menu.Selected >= len(menu.Items) {
		menu.Selected = 0
	}
}

// SelectRow updates hover row for list views.
func (s *MouseState) SelectRow(row int) {
	s.HoverRow = row
	s.Selection = SelectionRange{Start: row, End: row}
}

// ClearContextMenu closes any open menu.
func (s *MouseState) ClearContextMenu() {
	s.ContextMenu = nil
}

// ListRowFromY maps a terminal Y coordinate to a zero-based list row index.
func ListRowFromY(y, contentTopY int) int {
	if y < contentTopY {
		return -1
	}
	return y - contentTopY
}

// SelectedMenuItem returns the highlighted context menu item, if any.
func SelectedMenuItem(menu *ContextMenu) *MenuItem {
	if menu == nil || len(menu.Items) == 0 {
		return nil
	}
	idx := menu.Selected
	if idx < 0 {
		idx = 0
	}
	if idx >= len(menu.Items) {
		idx = len(menu.Items) - 1
	}
	item := menu.Items[idx]
	return &item
}
