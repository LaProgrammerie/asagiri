package renderers

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// Format selects an export format for graph visualization (spec-my-E §7).
type Format string

const (
	FormatJSON    Format = "json"
	FormatDOT     Format = "dot"
	FormatMermaid Format = "mermaid"
)

// Render exports the graph in the requested format deterministically.
func Render(graph knowledge.KnowledgeGraph, format Format) (string, error) {
	switch format {
	case FormatJSON:
		return RenderJSON(graph)
	case FormatDOT:
		return RenderDOT(graph), nil
	case FormatMermaid:
		return RenderMermaid(graph), nil
	default:
		return "", fmt.Errorf("unsupported render format %q", format)
	}
}
