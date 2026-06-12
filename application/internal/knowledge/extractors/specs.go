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

// SpecExtractor indexes Kiro specs and active AI docs (spec-my-E §10).
type SpecExtractor struct{}

func (SpecExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	var nodes []knowledge.GraphNode
	var warnings []string

	walk := func(root, kind string) error {
		return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") && !strings.HasSuffix(d.Name(), ".mdx") {
				return nil
			}
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			key := sanitizeStableKey(rel)
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   knowledge.NodeID(knowledge.NodeTypeFile, "spec_"+key),
				Type: knowledge.NodeTypeFile,
				Name: d.Name(),
				Path: rel,
				Properties: map[string]any{
					"artefact": "spec",
					"kind":     kind,
				},
				Source: knowledge.GraphSource{
					Kind:      kind,
					Path:      rel,
					Extractor: "specs",
				},
				Confidence: 0.85,
			}, now))
			return nil
		})
	}

	for _, dir := range []struct{ path, kind string }{
		{filepath.Join(repoRoot, ".kiro", "specs"), "kiro_spec"},
		{filepath.Join(repoRoot, "docs", "ai", "active"), "ai_active"},
	} {
		if err := walk(dir.path, dir.kind); err != nil && !os.IsNotExist(err) {
			return nil, nil, nil, err
		}
	}
	return nodes, nil, warnings, nil
}
