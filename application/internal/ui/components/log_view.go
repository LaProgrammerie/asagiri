package components

import (
	"fmt"
	"strings"
)

// LogLine is one log row.
type LogLine struct {
	Time    string
	Level   string
	Message string
}

// LogViewModel configures log rendering.
type LogViewModel struct {
	Lines   []LogLine
	Cursor  int
	Focused bool
	Limit   int
}

// RenderLogView renders log lines with virtual windowing.
func RenderLogView(vm LogViewModel) string {
	if len(vm.Lines) == 0 {
		return "No log lines"
	}
	visible := vm.Limit
	if visible <= 0 {
		visible = 12
	}
	win := SliceWindow(len(vm.Lines), vm.Cursor, visible)
	lines := VisibleSlice(vm.Lines, win)
	var b strings.Builder
	for i, line := range lines {
		idx := win.Offset + i
		prefix := " "
		if vm.Focused && idx == vm.Cursor {
			prefix = ">"
		}
		level := strings.TrimSpace(line.Level)
		if level == "" {
			level = "info"
		}
		glyph := StatusGlyph(ParseVisualState(level), false)
		ts := line.Time
		if ts == "" {
			ts = "--:--:--"
		}
		fmt.Fprintf(&b, "%s %s %s %s\n", prefix, ts, glyph, line.Message)
	}
	if win.Total > win.Limit {
		fmt.Fprintf(&b, "… showing %d-%d of %d", win.Offset+1, win.Offset+len(lines), win.Total)
	}
	return strings.TrimRight(b.String(), "\n")
}
