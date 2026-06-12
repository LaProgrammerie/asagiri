package renderers

import "github.com/LaProgrammerie/asagiri/application/internal/knowledge"

// RenderJSON exports the graph as indented JSON with sorted nodes and edges.
func RenderJSON(graph knowledge.KnowledgeGraph) (string, error) {
	return graph.MarshalExportJSON()
}
