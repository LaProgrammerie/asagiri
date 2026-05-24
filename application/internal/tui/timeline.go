package tui

import (
	"fmt"
	"io"
	"time"
)

// TimelineEntry is one finished step with duration.
type TimelineEntry struct {
	Label string
	Done  bool
	Since time.Duration
}

// RenderTimeline prints a simple checklist.
func RenderTimeline(out io.Writer, entries []TimelineEntry) {
	for _, e := range entries {
		mark := "⠋"
		if e.Done {
			mark = "✓"
		}
		fmt.Fprintf(out, "%s %-30s %s\n", mark, e.Label, formatSince(e.Since))
	}
}

func formatSince(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	return d.Round(time.Millisecond).String()
}
