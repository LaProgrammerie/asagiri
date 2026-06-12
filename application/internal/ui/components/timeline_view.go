package components

import (
	"fmt"
	"strings"
	"time"
)

// TimelineEntry is one timeline row.
type TimelineEntry struct {
	Time     time.Time
	Label    string
	Artifact string
}

// TimelineViewModel configures timeline rendering.
type TimelineViewModel struct {
	Title   string
	Entries []TimelineEntry
	Cursor  int
	Focused bool
}

// RenderTimelineView renders a selectable timeline.
func RenderTimelineView(vm TimelineViewModel) string {
	var b strings.Builder
	if vm.Title != "" {
		b.WriteString(vm.Title + "\n")
	}
	if len(vm.Entries) == 0 {
		b.WriteString("- none")
		return strings.TrimRight(b.String(), "\n")
	}
	for i, entry := range vm.Entries {
		prefix := " "
		if vm.Focused && i == vm.Cursor {
			prefix = ">"
		}
		label := entry.Label
		if entry.Artifact != "" {
			label += " artifact=" + entry.Artifact
		}
		ts := entry.Time.Format("15:04:05")
		if entry.Time.IsZero() {
			ts = "--:--:--"
		}
		fmt.Fprintf(&b, "%s %s %s\n", prefix, ts, label)
	}
	return strings.TrimRight(b.String(), "\n")
}
