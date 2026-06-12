package flows

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
	"gopkg.in/yaml.v3"
)

type flowDoc struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	Steps []struct {
		ID     string `yaml:"id"`
		Screen string `yaml:"screen"`
		Next   string `yaml:"next"`
	} `yaml:"steps"`
}

// Build reads product flow YAML files under flowsDir.
func Build(flowsDir string) (graph.Graph, error) {
	g := graph.Graph{Kind: "flow"}
	entries, err := os.ReadDir(flowsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return g, nil
		}
		return g, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(flowsDir, e.Name()))
		if err != nil {
			continue
		}
		var doc flowDoc
		if err := yaml.Unmarshal(raw, &doc); err != nil || doc.ID == "" {
			continue
		}
		fid := "flow:" + doc.ID
		g.Nodes = append(g.Nodes, graph.Node{ID: fid, Kind: "flow", Name: doc.Title})
		var prev string
		for _, step := range doc.Steps {
			sid := "step:" + doc.ID + "/" + step.ID
			g.Nodes = append(g.Nodes, graph.Node{ID: sid, Kind: "step", Name: step.ID})
			g.Edges = append(g.Edges, graph.Edge{From: fid, To: sid, Kind: "contains"})
			if step.Screen != "" {
				scr := "screen:" + step.Screen
				g.Nodes = append(g.Nodes, graph.Node{ID: scr, Kind: "screen", Name: step.Screen})
				g.Edges = append(g.Edges, graph.Edge{From: sid, To: scr, Kind: "uses_screen"})
			}
			if prev != "" {
				g.Edges = append(g.Edges, graph.Edge{From: prev, To: sid, Kind: "next"})
			}
			prev = sid
		}
	}
	return g, nil
}
