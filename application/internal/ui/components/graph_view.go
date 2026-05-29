package components

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// GraphViewModel configures graph rendering.
type GraphViewModel struct {
	Nodes   []bus.GraphNodeSummary
	Cursor  int
	Focused bool
	Limit   int
}

// RenderGraphView renders a compact node list.
func RenderGraphView(vm GraphViewModel) string {
	if len(vm.Nodes) == 0 {
		return "No graph nodes"
	}
	limit := vm.Limit
	if limit <= 0 {
		limit = 8
	}
	var b strings.Builder
	for i, node := range vm.Nodes {
		if i >= limit {
			b.WriteString(fmt.Sprintf("… +%d more\n", len(vm.Nodes)-limit))
			break
		}
		prefix := " "
		if vm.Focused && i == vm.Cursor {
			prefix = ">"
		}
		b.WriteString(prefix + GraphNodeCard(node, false) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
