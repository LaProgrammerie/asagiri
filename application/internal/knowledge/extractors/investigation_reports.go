package extractors

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// InvestigationReportExtractor indexes .asagiri/investigations (spec-my-E §10).
type InvestigationReportExtractor struct{}

func (InvestigationReportExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	dir := filepath.Join(repoRoot, ".asagiri", "investigations")
	var nodes []knowledge.GraphNode
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "report.md" && d.Name() != "graph.json" {
			return nil
		}
		invID := filepath.Base(filepath.Dir(path))
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:         knowledge.NodeID(knowledge.NodeTypeIncident, sanitizeStableKey(invID)),
			Type:       knowledge.NodeTypeIncident,
			Name:       invID,
			Path:       rel,
			Properties: map[string]any{"report": d.Name()},
			Source: knowledge.GraphSource{
				Kind:      "investigation",
				Path:      rel,
				Extractor: "investigation_reports",
			},
			Confidence: 0.83,
		}, now))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, nil, err
	}
	return nodes, nil, nil, nil
}
