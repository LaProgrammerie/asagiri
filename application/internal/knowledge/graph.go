package knowledge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// KnowledgeGraph is an in-memory graph snapshot (spec-my-E §22).
type KnowledgeGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// Validate checks structural integrity of the graph.
func (g KnowledgeGraph) Validate() error {
	nodeIDs := make(map[string]struct{}, len(g.Nodes))
	for _, n := range g.Nodes {
		if err := n.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidGraph, err)
		}
		if _, dup := nodeIDs[n.ID]; dup {
			return fmt.Errorf("%w: duplicate node id %q", ErrInvalidGraph, n.ID)
		}
		nodeIDs[n.ID] = struct{}{}
	}

	edgeIDs := make(map[string]struct{}, len(g.Edges))
	for _, e := range g.Edges {
		if err := e.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidGraph, err)
		}
		if _, dup := edgeIDs[e.ID]; dup {
			return fmt.Errorf("%w: duplicate edge id %q", ErrInvalidGraph, e.ID)
		}
		edgeIDs[e.ID] = struct{}{}
		if _, ok := nodeIDs[e.From]; !ok {
			return fmt.Errorf("%w: edge from unknown node %q", ErrInvalidGraph, e.From)
		}
		if _, ok := nodeIDs[e.To]; !ok {
			return fmt.Errorf("%w: edge to unknown node %q", ErrInvalidGraph, e.To)
		}
	}
	return nil
}

// PruneOrphanEdges drops edges whose endpoints are missing from nodes.
// Returns the pruned graph and orphan edge ids (for build warnings).
func PruneOrphanEdges(g KnowledgeGraph) (KnowledgeGraph, []string) {
	nodeIDs := make(map[string]struct{}, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeIDs[n.ID] = struct{}{}
	}
	var kept []GraphEdge
	var dropped []string
	for _, e := range g.Edges {
		if _, ok := nodeIDs[e.From]; !ok {
			dropped = append(dropped, e.ID)
			continue
		}
		if _, ok := nodeIDs[e.To]; !ok {
			dropped = append(dropped, e.ID)
			continue
		}
		kept = append(kept, e)
	}
	if len(dropped) == 0 {
		return g, nil
	}
	out := g
	out.Edges = kept
	return out, dropped
}

// SortedNodes returns nodes ordered by ID for deterministic rendering.
func (g KnowledgeGraph) SortedNodes() []GraphNode {
	out := append([]GraphNode(nil), g.Nodes...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// SortedEdges returns edges ordered by id.
func (g KnowledgeGraph) SortedEdges() []GraphEdge {
	out := append([]GraphEdge(nil), g.Edges...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// MarshalExportJSON returns indented JSON with sorted nodes and edges.
func (g KnowledgeGraph) MarshalExportJSON() (string, error) {
	export := KnowledgeGraph{
		Nodes: g.SortedNodes(),
		Edges: g.SortedEdges(),
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(export); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// ParseJSON unmarshals a knowledge graph from JSON bytes.
func ParseJSON(body []byte) (KnowledgeGraph, error) {
	var graph KnowledgeGraph
	if err := json.Unmarshal(body, &graph); err != nil {
		return KnowledgeGraph{}, fmt.Errorf("parse knowledge graph json: %w", err)
	}
	return graph, nil
}
