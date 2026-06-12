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

// InfraExtractor scans infrastructure/ for Terraform and Docker resources (light V1).
type InfraExtractor struct{}

func (InfraExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	root := filepath.Join(repoRoot, "infrastructure")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil, nil, nil
	}
	var nodes []knowledge.GraphNode
	seen := map[string]struct{}{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".terraform" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !isInfraResourceFile(d.Name()) {
			return nil
		}
		key := sanitizeStableKey(rel)
		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}
		kind := infraKind(d.Name())
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   knowledge.NodeID(knowledge.NodeTypeInfraResource, key),
			Type: knowledge.NodeTypeInfraResource,
			Name: filepath.Base(path),
			Path: rel,
			Properties: map[string]any{
				"kind": kind,
			},
			Source: knowledge.GraphSource{
				Kind:      "infra",
				Path:      rel,
				Extractor: "infra",
			},
			Confidence: confContractMid,
		}, now))
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}
	return nodes, nil, nil, nil
}

func isInfraResourceFile(name string) bool {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tf"),
		strings.HasSuffix(lower, ".tf.json"),
		lower == "dockerfile",
		strings.HasPrefix(lower, "docker-compose"),
		strings.HasSuffix(lower, "compose.yaml"),
		strings.HasSuffix(lower, "compose.yml"):
		return true
	default:
		return false
	}
}

func infraKind(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".tf"), strings.HasSuffix(lower, ".tf.json"):
		return "terraform"
	case lower == "dockerfile" || strings.Contains(lower, "compose"):
		return "docker"
	default:
		return "infra"
	}
}
