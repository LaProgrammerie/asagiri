package knowledge

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// PersistGraphJSON writes graph as JSON under repoRoot/.asagiri/knowledge/graph.json.
func PersistGraphJSON(repoRoot string, g KnowledgeGraph) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("persist knowledge graph json: %w", err)
	}
	body, err := g.MarshalExportJSON()
	if err != nil {
		return fmt.Errorf("persist knowledge graph json: render: %w", err)
	}
	dir := filepath.Join(repoRoot, KnowledgeRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("persist knowledge graph json: mkdir: %w", err)
	}
	path := filepath.Join(dir, GraphJSONName)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("persist knowledge graph json: write: %w", err)
	}
	return nil
}

// ExportRepoGraph loads the SQLite store and writes graph.json (spec-my-E §5).
func ExportRepoGraph(repoRoot string) error {
	store, err := OpenStore(repoRoot)
	if err != nil {
		return fmt.Errorf("export repo knowledge graph: %w", err)
	}
	defer store.Close()

	graph, err := store.LoadGraph(context.Background())
	if err != nil {
		return fmt.Errorf("export repo knowledge graph: load: %w", err)
	}
	if err := PersistGraphJSON(repoRoot, graph); err != nil {
		return fmt.Errorf("export repo knowledge graph: %w", err)
	}
	return nil
}
