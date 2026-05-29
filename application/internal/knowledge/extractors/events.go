package extractors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"gopkg.in/yaml.v3"
)

// EventExtractor parses domain events from contracts/events.yaml.
type EventExtractor struct{}

func (EventExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	rel := relProductPath(product, "contracts", "events.yaml")
	abs := filepath.Join(productDir(repoRoot, product), "contracts", "events.yaml")
	body, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	names, warnings, err := parseEventNames(body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("events: %w", err)
	}
	if len(names) == 0 {
		return nil, nil, warnings, nil
	}

	productID := knowledge.NodeID(knowledge.NodeTypeProduct, product)
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge

	for _, name := range names {
		eventID := knowledge.NodeID(knowledge.NodeTypeEvent, sanitizeStableKey(name))
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   eventID,
			Type: knowledge.NodeTypeEvent,
			Name: name,
			Path: rel,
			Source: knowledge.GraphSource{
				Kind:      "events",
				Path:      rel,
				Extractor: "events",
			},
			Confidence: confEventHigh,
		}, now))
		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, productID, eventID),
			From: productID,
			To:   eventID,
			Type: knowledge.EdgeTypeOwns,
			Source: knowledge.GraphSource{
				Kind:      "events",
				Path:      rel,
				Extractor: "events",
			},
			Confidence: confContractMid,
		}, now))
	}
	return nodes, edges, warnings, nil
}

func parseEventNames(body []byte) ([]string, []string, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, nil, err
	}
	var warnings []string
	var names []string

	switch events := doc["events"].(type) {
	case []any:
		for _, item := range events {
			switch v := item.(type) {
			case string:
				if v != "" {
					names = append(names, v)
				}
			case map[string]any:
				if n, _ := v["name"].(string); n != "" {
					names = append(names, n)
				} else if n, _ := v["id"].(string); n != "" {
					names = append(names, n)
				}
			}
		}
	case map[string]any:
		for k := range events {
			names = append(names, k)
		}
	default:
		warnings = append(warnings, "events.yaml: no events section found")
	}
	return names, warnings, nil
}

func sanitizeStableKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}
