package extractors

import (
	"context"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

const runtimeEventLimit = 200

// RuntimeExtractor reads .asagiri/runtime events when present (best effort).
type RuntimeExtractor struct{}

func (RuntimeExtractor) Extract(ctx context.Context, repoRoot, _ string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	store, err := runtime.Open(repoRoot)
	if err != nil {
		return nil, nil, []string{"runtime: " + err.Error()}, nil
	}
	defer func() { _ = store.Close() }()

	events, err := store.ListEvents(runtimeEventLimit)
	if err != nil {
		return nil, nil, []string{"runtime: " + err.Error()}, nil
	}
	if len(events) == 0 {
		return nil, nil, nil, nil
	}

	now := time.Now().UTC()
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	seen := map[string]struct{}{}

	for _, ev := range events {
		evType := strings.TrimSpace(ev.Type)
		if evType == "" {
			continue
		}
		stableKey := sanitizeStableKey(evType)
		if _, ok := seen[stableKey]; ok {
			continue
		}
		seen[stableKey] = struct{}{}
		eventID := knowledge.NodeID(knowledge.NodeTypeEvent, "runtime_"+stableKey)
		rel := ".asagiri/runtime"
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   eventID,
			Type: knowledge.NodeTypeEvent,
			Name: evType,
			Path: rel,
			Properties: map[string]any{
				"source": "runtime",
			},
			Source: knowledge.GraphSource{
				Kind:      "runtime",
				Path:      rel,
				Extractor: "runtime",
				Evidence:  ev.Source,
			},
			Confidence: confEventHigh * 0.85,
		}, now))

		if flowID := strings.TrimSpace(ev.FlowID); flowID != "" {
			flowNodeID := knowledge.NodeID(knowledge.NodeTypeFlow, flowID)
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeProduces, flowNodeID, eventID),
				From: flowNodeID,
				To:   eventID,
				Type: knowledge.EdgeTypeProduces,
				Source: knowledge.GraphSource{
					Kind:      "runtime",
					Path:      rel,
					Extractor: "runtime",
				},
				Confidence: confEventHigh * 0.8,
			}, now))
		}
	}
	return nodes, edges, nil, nil
}
