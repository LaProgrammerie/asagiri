package knowledge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MarshalGraphJSON returns indented JSON for a knowledge graph snapshot.
func MarshalGraphJSON(graph KnowledgeGraph) (string, error) {
	export := KnowledgeGraph{
		Nodes: graph.SortedNodes(),
		Edges: graph.SortedEdges(),
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(export); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

func writeGraphJSON(repoRoot string, graph KnowledgeGraph) error {
	dir := filepath.Join(repoRoot, KnowledgeRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("knowledge export dir: %w", err)
	}
	body, err := MarshalGraphJSON(graph)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, graphJSONName)
	if err := os.WriteFile(path, []byte(body+"\n"), 0o644); err != nil {
		return fmt.Errorf("write graph.json: %w", err)
	}
	return nil
}
