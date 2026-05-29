package app

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
)

func logLinesFromEvents(events []bus.EventSummary) []components.LogLine {
	if len(events) == 0 {
		return nil
	}
	lines := make([]components.LogLine, 0, len(events))
	for _, ev := range events {
		level := "info"
		typ := strings.ToLower(ev.Type)
		switch {
		case strings.Contains(typ, "error"), strings.Contains(typ, "failed"):
			level = "error"
		case strings.Contains(typ, "warn"):
			level = "warning"
		case strings.Contains(typ, "completed"), strings.Contains(typ, "success"):
			level = "success"
		}
		msg := ev.Type
		if ev.Source != "" {
			msg += " (" + ev.Source + ")"
		}
		lines = append(lines, components.LogLine{
			Time:    ev.CreatedAt.Format("15:04:05"),
			Level:   level,
			Message: msg,
		})
	}
	return lines
}
