package extractors

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

var adrFilePattern = regexp.MustCompile(`^(\d{3})-([a-z0-9-]+)\.md$`)

// ADRExtractor scans docs/decisions/*.md for ADR nodes (spec-my-E §11).
type ADRExtractor struct{}

func (ADRExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	dir := filepath.Join(repoRoot, "docs", "decisions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	var nodes []knowledge.GraphNode
	var warnings []string
	relDir := "docs/decisions"

	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		if ent.Name() == "README.md" {
			continue
		}
		m := adrFilePattern.FindStringSubmatch(ent.Name())
		if m == nil {
			warnings = append(warnings, "adr: skipped non-ADR file "+ent.Name())
			continue
		}
		num, slug := m[1], m[2]
		rel := filepath.ToSlash(filepath.Join(relDir, ent.Name()))
		title := readADRTitle(filepath.Join(dir, ent.Name()), slug)
		stableKey := num + "_" + sanitizeStableKey(slug)
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   knowledge.NodeID(knowledge.NodeTypeADR, stableKey),
			Type: knowledge.NodeTypeADR,
			Name: title,
			Path: rel,
			Properties: map[string]any{
				"number": num,
				"slug":   slug,
			},
			Source: knowledge.GraphSource{
				Kind:      "adr",
				Path:      rel,
				Extractor: "adr",
			},
			Confidence: confContractHigh,
		}, now))
	}
	return nodes, nil, warnings, nil
}

var adrTitlePattern = regexp.MustCompile(`(?m)^#\s+ADR-\d+\s+[—–-]\s+(.+)$`)

func readADRTitle(absPath, fallback string) string {
	body, err := os.ReadFile(absPath)
	if err != nil {
		return fallback
	}
	if m := adrTitlePattern.FindSubmatch(body); len(m) > 1 {
		return strings.TrimSpace(string(m[1]))
	}
	lines := strings.SplitN(string(body), "\n", 5)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return fallback
}
