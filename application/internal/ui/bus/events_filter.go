package bus

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// EventFeedFilterTypes are the canonical event feed categories (spec-ui §22).
var EventFeedFilterTypes = []string{
	"all",
	"runtime",
	"agent",
	"trust",
	"investigation",
	"graph",
	"knowledge",
	"replay",
	"prototype",
}

func eventMatchesCategory(ev EventSummary, filter string) bool {
	f := strings.ToLower(strings.TrimSpace(filter))
	if f == "" || f == "all" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(ev.Type), f+".") || strings.Contains(strings.ToLower(ev.Type), f)
}

func eventMatchesSearchText(ev EventSummary, search string) bool {
	s := strings.ToLower(strings.TrimSpace(search))
	if s == "" || s == "(none)" {
		return true
	}
	return strings.Contains(strings.ToLower(ev.Type), s) ||
		strings.Contains(strings.ToLower(ev.Source), s) ||
		strings.Contains(strings.ToLower(ev.SessionID), s) ||
		strings.Contains(strings.ToLower(ev.FlowID), s) ||
		strings.Contains(strings.ToLower(ev.ID), s)
}

func mapRuntimeEvent(row runtime.RuntimeEvent) EventSummary {
	return EventSummary{
		ID:        row.ID,
		Type:      row.Type,
		Source:    row.Source,
		SessionID: row.SessionID,
		FlowID:    row.FlowID,
		CreatedAt: row.CreatedAt,
		Payload:   row.Payload,
	}
}
