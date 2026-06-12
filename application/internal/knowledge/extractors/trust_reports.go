package extractors

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// TrustReportExtractor indexes .asagiri/trust reports (spec-my-E §10).
type TrustReportExtractor struct{}

func (TrustReportExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	dir := filepath.Join(repoRoot, ".asagiri", "trust")
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") && !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		id := filepath.Base(filepath.Dir(path))
		if id == "trust" {
			id = strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		}
		nodeID := knowledge.NodeID(knowledge.NodeTypeReview, sanitizeStableKey("trust_"+id))
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   nodeID,
			Type: knowledge.NodeTypeReview,
			Name: id,
			Path: rel,
			Source: knowledge.GraphSource{
				Kind:      "trust_report",
				Path:      rel,
				Extractor: "trust_reports",
			},
			Confidence: 0.84,
		}, now))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, nil, err
	}
	return nodes, edges, nil, nil
}
