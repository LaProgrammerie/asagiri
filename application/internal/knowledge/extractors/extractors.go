package extractors

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// ExtractResult holds nodes, edges, and non-fatal warnings from an extractor run.
type ExtractResult struct {
	Nodes    []knowledge.GraphNode
	Edges    []knowledge.GraphEdge
	Warnings []string
}

const (
	confFlowHigh      = 0.95
	confFlowMid       = 0.92
	confContractHigh  = 0.94
	confContractMid   = 0.90
	confEventHigh     = 0.92
	confPermissionHigh = 0.91
)

func productDir(repoRoot, product string) string {
	return filepath.Join(repoRoot, ".asagiri", "products", product)
}

func relProductPath(product string, parts ...string) string {
	all := append([]string{".asagiri", "products", product}, parts...)
	return filepath.ToSlash(filepath.Join(all...))
}

func stampNode(node knowledge.GraphNode, now time.Time) knowledge.GraphNode {
	if node.CreatedAt.IsZero() {
		node.CreatedAt = now
	}
	node.UpdatedAt = now
	return node
}

func stampEdge(edge knowledge.GraphEdge, now time.Time) knowledge.GraphEdge {
	if edge.CreatedAt.IsZero() {
		edge.CreatedAt = now
	}
	edge.UpdatedAt = now
	return edge
}

// APIOperationKey builds a stable api_operation id segment from HTTP method and path.
func APIOperationKey(method, path string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")
	if path == "" {
		return method
	}
	return method + "_" + path
}

func apiOperationNode(method, path, relPath, extractor string, confidence float64, now time.Time) knowledge.GraphNode {
	key := APIOperationKey(method, path)
	name := strings.ToUpper(method) + " " + path
	return stampNode(knowledge.GraphNode{
		ID:         knowledge.NodeID(knowledge.NodeTypeAPIOperation, key),
		Type:       knowledge.NodeTypeAPIOperation,
		Name:       name,
		Path:       relPath,
		Source: knowledge.GraphSource{
			Kind:      "contract",
			Path:      relPath,
			Extractor: extractor,
		},
		Confidence: confidence,
	}, now)
}

func parseContractRef(ref string) (method, path string, ok bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(strings.ToUpper(ref), "TODO:") {
		return "", "", false
	}
	parts := strings.Fields(ref)
	if len(parts) < 2 {
		return "", "", false
	}
	return strings.ToUpper(parts[0]), parts[1], true
}
