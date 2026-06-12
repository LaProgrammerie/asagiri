package executiongraph

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RenderFormat selects an export format for graph visualization (spec §5.6).
type RenderFormat string

const (
	RenderFormatMermaid  RenderFormat = "mermaid"
	RenderFormatDOT      RenderFormat = "dot"
	RenderFormatMarkdown RenderFormat = "markdown"
	RenderFormatJSON     RenderFormat = "json"
)

// Render exports the graph in the requested format deterministically.
func Render(graph ExecutionGraph, format RenderFormat) (string, error) {
	switch format {
	case RenderFormatMermaid:
		return renderMermaid(graph), nil
	case RenderFormatDOT:
		return renderDOT(graph), nil
	case RenderFormatMarkdown:
		return renderMarkdown(graph), nil
	case RenderFormatJSON:
		return renderJSON(graph)
	default:
		return "", fmt.Errorf("unsupported render format %q", format)
	}
}

func renderMermaid(graph ExecutionGraph) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")
	for _, n := range graph.SortedNodes() {
		label := escapeMermaidLabel(n.ID)
		if n.Title != "" {
			label = escapeMermaidLabel(n.Title)
		}
		fmt.Fprintf(&sb, "  %s[%s]\n", sanitizeMermaidID(n.ID), label)
	}
	for _, e := range graph.SortedEdges() {
		fmt.Fprintf(&sb, "  %s -->|%s| %s\n",
			sanitizeMermaidID(e.From), e.Type, sanitizeMermaidID(e.To))
	}
	return sb.String()
}

func renderDOT(graph ExecutionGraph) string {
	var sb strings.Builder
	sb.WriteString("digraph execution_graph {\n")
	sb.WriteString("  rankdir=LR;\n")
	for _, n := range graph.SortedNodes() {
		label := n.ID
		if n.Title != "" {
			label = n.Title
		}
		fmt.Fprintf(&sb, "  %q [label=%q];\n", n.ID, dotEscape(label))
	}
	for _, e := range graph.SortedEdges() {
		fmt.Fprintf(&sb, "  %q -> %q [label=%q];\n", e.From, e.To, string(e.Type))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func renderMarkdown(graph ExecutionGraph) string {
	var sb strings.Builder
	sb.WriteString("# Execution Graph\n\n")
	fmt.Fprintf(&sb, "- ID: `%s`\n", graph.ID)
	fmt.Fprintf(&sb, "- Product: `%s`\n", graph.Product)
	if graph.Flow != "" {
		fmt.Fprintf(&sb, "- Flow: `%s`\n", graph.Flow)
	}
	fmt.Fprintf(&sb, "- Status: `%s`\n", graph.Status)
	fmt.Fprintf(&sb, "- Max parallel: `%d`\n\n", graph.Strategy.MaxParallel)

	sb.WriteString("## Nodes\n\n")
	for _, n := range graph.SortedNodes() {
		fmt.Fprintf(&sb, "- `%s` (%s)", n.ID, n.Type)
		if n.Title != "" {
			fmt.Fprintf(&sb, ": %s", n.Title)
		}
		if n.Risk != "" {
			fmt.Fprintf(&sb, " [risk=%s]", n.Risk)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n## Edges\n\n")
	for _, e := range graph.SortedEdges() {
		line := fmt.Sprintf("- `%s` → `%s` (%s)", e.From, e.To, e.Type)
		if e.Reason != "" {
			line += fmt.Sprintf(": %s", e.Reason)
		}
		sb.WriteString(line + "\n")
	}

	if len(graph.Checkpoints) > 0 {
		sb.WriteString("\n## Checkpoints\n\n")
		for _, cp := range graph.Checkpoints {
			fmt.Fprintf(&sb, "- after `%s`\n", cp.After)
		}
	}

	if graph.Rollback != nil {
		sb.WriteString("\n## Rollback\n\n")
		fmt.Fprintf(&sb, "- Strategy: `%s`\n", graph.Rollback.Strategy)
		if graph.Rollback.PreserveReports {
			sb.WriteString("- Preserve reports: `true`\n")
		}
	}

	return sb.String()
}

func renderJSON(graph ExecutionGraph) (string, error) {
	body, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// RenderPlanMD renders the plan.md skeleton for a graph (spec §23).
func RenderPlanMD(graph ExecutionGraph) string {
	var sb strings.Builder
	sb.WriteString("# Execution Plan\n\n")
	fmt.Fprintf(&sb, "- Graph ID: `%s`\n", graph.ID)
	fmt.Fprintf(&sb, "- Product: `%s`\n", graph.Product)
	if graph.Flow != "" {
		fmt.Fprintf(&sb, "- Flow: `%s`\n", graph.Flow)
	}
	fmt.Fprintf(&sb, "- Status: `%s`\n", graph.Status)
	fmt.Fprintf(&sb, "- Created: `%s`\n", graph.CreatedAt)
	fmt.Fprintf(&sb, "- Nodes: %d\n", len(graph.Nodes))
	fmt.Fprintf(&sb, "- Edges: %d\n", len(graph.Edges))
	fmt.Fprintf(&sb, "- Checkpoints: %d\n", len(graph.Checkpoints))
	fmt.Fprintf(&sb, "- Max parallel: %d\n\n", graph.Strategy.MaxParallel)

	sb.WriteString("## Summary\n\n")
	sb.WriteString("_Planning decisions will be recorded here._\n\n")

	sb.WriteString("## Graph\n\n")
	sb.WriteString("```mermaid\n")
	sb.WriteString(renderMermaid(graph))
	sb.WriteString("```\n")

	return sb.String()
}

func sanitizeMermaidID(id string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(id)
}

func escapeMermaidLabel(s string) string {
	return strings.NewReplacer("\n", " ", `"`, "").Replace(s)
}

func dotEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`)
}
