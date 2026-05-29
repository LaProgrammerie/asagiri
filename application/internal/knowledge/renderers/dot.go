package renderers

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// RenderDOT exports the graph as Graphviz DOT.
func RenderDOT(graph knowledge.KnowledgeGraph) string {
	var sb strings.Builder
	sb.WriteString("digraph knowledge_graph {\n")
	sb.WriteString("  rankdir=LR;\n")
	for _, n := range graph.SortedNodes() {
		label := nodeLabel(n)
		sb.WriteString(fmt.Sprintf("  %q [label=%q];\n", n.ID, dotEscape(label)))
	}
	for _, e := range graph.SortedEdges() {
		label := edgeLabel(e)
		sb.WriteString(fmt.Sprintf("  %q -> %q [label=%q];\n", e.From, e.To, dotEscape(label)))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func nodeLabel(n knowledge.GraphNode) string {
	name := n.ID
	if n.Name != "" {
		name = n.Name
	}
	if n.Type != "" {
		return string(n.Type) + "\\n" + name
	}
	return name
}

func edgeLabel(e knowledge.GraphEdge) string {
	if e.Confidence > 0 {
		return fmt.Sprintf("%s (%.2f)", e.Type, e.Confidence)
	}
	return string(e.Type)
}

func dotEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
}
