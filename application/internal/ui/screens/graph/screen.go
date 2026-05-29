package graph

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/components"
)

// ViewModel contains read-only graph explorer data.
type ViewModel struct {
	Graph   bus.GraphExplorerResult
	Events  []bus.EventSummary
	ShowCLI bool
}

// Render returns the graph explorer textual view.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Graph: %s\n", empty(vm.Graph.GraphID, "-")))
	b.WriteString(fmt.Sprintf("Flow: %s\n", empty(vm.Graph.FlowID, "-")))
	b.WriteString(fmt.Sprintf("Status: %s\n", empty(vm.Graph.Status, "unknown")))
	b.WriteString("\nNodes\n")
	if len(vm.Graph.Nodes) == 0 {
		b.WriteString("- none\n")
	} else {
		for i, node := range vm.Graph.Nodes {
			if i >= 8 {
				break
			}
			b.WriteString(fmt.Sprintf("- %s [%s] %s risk=%s\n", empty(node.ID, "-"), empty(node.Status, "unknown"), empty(node.Title, "-"), empty(node.Risk, "unknown")))
			if len(node.BlockedBy) > 0 {
				b.WriteString("  blocked by: " + strings.Join(node.BlockedBy, ", ") + "\n")
			}
			if vm.ShowCLI && node.CLIEquivalent != "" {
				b.WriteString("  CLI: " + node.CLIEquivalent + "\n")
			}
		}
	}
	b.WriteString("\nActions\n")
	b.WriteString("- open node (stub)\n")
	b.WriteString("- view dependencies (stub)\n")
	b.WriteString("- export mermaid/json: asa graph visualize <graph-id> --format mermaid\n")
	b.WriteString("\nEvents\n")
	b.WriteString(components.RenderEventFeed(components.EventFeedViewModel{
		Events:       vm.Events,
		Limit:        5,
		ShowCLIHints: vm.ShowCLI,
	}))
	return strings.TrimRight(b.String(), "\n")
}

func empty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
