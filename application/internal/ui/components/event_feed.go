package components

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// EventFeedViewModel configures reusable event feed rendering.
type EventFeedViewModel struct {
	Events       []bus.EventSummary
	Limit        int
	Filter       string
	Search       string
	ShowCLIHints bool
}

// RenderEventFeed renders a reusable event feed block with filter/search stubs.
func RenderEventFeed(vm EventFeedViewModel) string {
	limit := vm.Limit
	if limit <= 0 {
		limit = 5
	}
	filter := strings.TrimSpace(vm.Filter)
	if filter == "" {
		filter = "all (stub)"
	}
	search := strings.TrimSpace(vm.Search)
	if search == "" {
		search = "(none)"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Filter: %s  Search: %s\n", filter, search))
	if vm.ShowCLIHints {
		b.WriteString("CLI: asa runtime events --type <filter> --search <query>\n")
	}

	if len(vm.Events) == 0 {
		b.WriteString("- none")
		return b.String()
	}

	count := 0
	for _, ev := range vm.Events {
		if count >= limit {
			break
		}
		if !eventMatchesFilter(ev, filter) || !eventMatchesSearch(ev, search) {
			continue
		}
		b.WriteString(fmt.Sprintf("- %s  %s\n", ev.CreatedAt.Format("15:04:05"), ev.Type))
		count++
	}
	if count == 0 {
		b.WriteString("- none")
	}
	return strings.TrimRight(b.String(), "\n")
}

func eventMatchesFilter(ev bus.EventSummary, filter string) bool {
	f := strings.ToLower(strings.TrimSpace(filter))
	if f == "" || f == "all (stub)" || f == "all" {
		return true
	}
	return strings.Contains(strings.ToLower(ev.Type), f)
}

func eventMatchesSearch(ev bus.EventSummary, search string) bool {
	s := strings.ToLower(strings.TrimSpace(search))
	if s == "" || s == "(none)" {
		return true
	}
	return strings.Contains(strings.ToLower(ev.Type), s) || strings.Contains(strings.ToLower(ev.Source), s)
}
