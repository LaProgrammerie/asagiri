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

// TaskExtractor indexes .asagiri/tasks artefacts (spec-my-E §10).
type TaskExtractor struct{}

func (TaskExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	dir := filepath.Join(repoRoot, ".asagiri", "tasks")
	var nodes []knowledge.GraphNode
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		key := sanitizeStableKey(rel)
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   knowledge.NodeID(knowledge.NodeTypeFile, "task_"+key),
			Type: knowledge.NodeTypeFile,
			Name: strings.TrimSuffix(d.Name(), ".yaml"),
			Path: rel,
			Properties: map[string]any{"artefact": "task"},
			Source: knowledge.GraphSource{
				Kind:      "task",
				Path:      rel,
				Extractor: "tasks",
			},
			Confidence: 0.86,
		}, now))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, nil, err
	}
	return nodes, nil, nil, nil
}
