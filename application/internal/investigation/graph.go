package investigation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GraphNode is a node in the root-cause graph (spec-my-A §25.13).
type GraphNode struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Label string `json:"label"`
}

// GraphEdge links investigation artefacts.
type GraphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

// RootCauseGraph is the investigation graph export.
type RootCauseGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// BuildRootCauseGraph constructs a graph from a report and optional knowledge context pack.
func BuildRootCauseGraph(rep Report, pack ContextPack) RootCauseGraph {
	var g RootCauseGraph
	symptomID := "symptom:0"
	g.Nodes = append(g.Nodes, GraphNode{ID: symptomID, Kind: "symptom", Label: rep.Request.Symptom})
	if rep.Scope.Flow != "" {
		fid := "flow:" + rep.Scope.Flow
		g.Nodes = append(g.Nodes, GraphNode{ID: fid, Kind: "flow", Label: rep.Scope.Flow})
		g.Edges = append(g.Edges, GraphEdge{From: symptomID, To: fid, Relation: "fails_at"})
	}
	for i, api := range pack.APIs {
		nid := fmt.Sprintf("api:%d", i)
		g.Nodes = append(g.Nodes, GraphNode{ID: nid, Kind: "api", Label: api})
		if rep.Scope.Flow != "" {
			g.Edges = append(g.Edges, GraphEdge{From: "flow:" + rep.Scope.Flow, To: nid, Relation: "requires"})
		}
	}
	for i, ev := range pack.Events {
		nid := fmt.Sprintf("event:%d", i)
		g.Nodes = append(g.Nodes, GraphNode{ID: nid, Kind: "event", Label: ev})
		if rep.Scope.Flow != "" {
			g.Edges = append(g.Edges, GraphEdge{From: "flow:" + rep.Scope.Flow, To: nid, Relation: "emits"})
		}
	}
	for i, e := range rep.Evidence {
		nid := "evidence:" + e.ID
		if e.ID == "" {
			nid = fmt.Sprintf("evidence:%d", i)
		}
		g.Nodes = append(g.Nodes, GraphNode{ID: nid, Kind: string(e.Kind), Label: e.Summary})
		g.Edges = append(g.Edges, GraphEdge{From: symptomID, To: nid, Relation: "supports"})
	}
	for i, h := range rep.Hypotheses {
		hid := "hypothesis:" + h.ID
		if h.ID == "" {
			hid = fmt.Sprintf("hypothesis:%d", i)
		}
		g.Nodes = append(g.Nodes, GraphNode{ID: hid, Kind: "hypothesis", Label: h.Statement})
		g.Edges = append(g.Edges, GraphEdge{From: symptomID, To: hid, Relation: "supports"})
	}
	for _, f := range rep.LocalResult.CandidateFiles {
		fid := "file:" + f
		g.Nodes = append(g.Nodes, GraphNode{ID: fid, Kind: "file", Label: f})
		g.Edges = append(g.Edges, GraphEdge{From: symptomID, To: fid, Relation: "implemented_by"})
	}
	return g
}

// WriteGraph persists graph.json under the investigation directory.
func WriteGraph(repoRoot string, rep Report, pack ContextPack) (string, error) {
	dir := filepath.Join(repoRoot, ".asagiri", "investigations", rep.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	g, err := BuildRootCauseGraphWithKnowledge(context.Background(), repoRoot, rep, pack)
	if err != nil {
		g = BuildRootCauseGraph(rep, pack)
	}
	raw, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "graph.json")
	return path, os.WriteFile(path, raw, 0o644)
}

// FormatGraphPlain renders a minimal CLI visualization.
func FormatGraphPlain(g RootCauseGraph) string {
	var b strings.Builder
	b.WriteString("Root Cause Graph\n")
	b.WriteString("────────────────\n")
	for _, n := range g.Nodes {
		b.WriteString("  [" + n.Kind + "] " + n.Label + "\n")
	}
	for _, e := range g.Edges {
		b.WriteString("  " + e.From + " -" + e.Relation + "-> " + e.To + "\n")
	}
	return b.String()
}
