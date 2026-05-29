package extractors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// ConfigExtractor indexes Asagiri and application config files (spec-my-E §10).
type ConfigExtractor struct{}

func (ConfigExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	var nodes []knowledge.GraphNode
	var warnings []string

	candidates := []string{
		filepath.Join(repoRoot, ".asagiri", "config.yaml"),
		filepath.Join(repoRoot, ".asagiri", "config.yaml.example"),
	}
	appCfg := filepath.Join(repoRoot, "application", "config")
	if entries, err := os.ReadDir(appCfg); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
				candidates = append(candidates, filepath.Join(appCfg, e.Name()))
			}
		}
	}

	for _, abs := range candidates {
		if _, err := os.Stat(abs); err != nil {
			continue
		}
		rel, err := filepath.Rel(repoRoot, abs)
		if err != nil {
			continue
		}
		rel = filepath.ToSlash(rel)
		key := sanitizeStableKey(rel)
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   knowledge.NodeID(knowledge.NodeTypeConfig, key),
			Type: knowledge.NodeTypeConfig,
			Name: filepath.Base(abs),
			Path: rel,
			Source: knowledge.GraphSource{
				Kind:      "config",
				Path:      rel,
				Extractor: "config",
			},
			Confidence: 0.9,
		}, now))
	}
	return nodes, nil, warnings, nil
}
