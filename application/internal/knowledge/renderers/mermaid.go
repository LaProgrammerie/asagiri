package renderers

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// RenderMermaid exports the graph as a Mermaid flowchart.
func RenderMermaid(graph knowledge.KnowledgeGraph) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	for _, n := range graph.SortedNodes() {
		label := escapeMermaidLabel(nodeLabel(n))
		fmt.Fprintf(&sb, "  %s[%s]\n", sanitizeMermaidID(n.ID), label)
	}
	for _, e := range graph.SortedEdges() {
		fmt.Fprintf(&sb, "  %s -->|%s| %s\n",
			sanitizeMermaidID(e.From), escapeMermaidLabel(edgeLabel(e)), sanitizeMermaidID(e.To))
	}
	return sb.String()
}

func sanitizeMermaidID(id string) string {
	return strings.NewReplacer("-", "_", ".", "_", ":", "_").Replace(id)
}

func escapeMermaidLabel(s string) string {
	return strings.NewReplacer("\n", " ", `"`, "", "|", "/").Replace(s)
}
