package events

import (
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
	"gopkg.in/yaml.v3"
)

// Build reads events contract YAML (map name -> definition).
func Build(eventsPath string) (graph.Graph, error) {
	g := graph.Graph{Kind: "event"}
	raw, err := os.ReadFile(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return g, nil
		}
		return g, err
	}
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return g, err
	}
	events, _ := root["events"].(map[string]any)
	if events == nil {
		for k := range root {
			if k == "version" || k == "schema" {
				continue
			}
			ename := "event:" + k
			g.Nodes = append(g.Nodes, graph.Node{ID: ename, Kind: "event", Name: k})
		}
		return g, nil
	}
	for name := range events {
		ename := "event:" + name
		g.Nodes = append(g.Nodes, graph.Node{ID: ename, Kind: "event", Name: name})
	}
	// domain links
	if dom, ok := root["domains"].(map[string]any); ok {
		for domain, evs := range dom {
			did := "domain:" + domain
			g.Nodes = append(g.Nodes, graph.Node{ID: did, Kind: "domain", Name: domain})
			list, _ := evs.([]any)
			for _, item := range list {
				ename := "event:" + strings.TrimSpace(item.(string))
				g.Edges = append(g.Edges, graph.Edge{From: ename, To: did, Kind: "belongs_to"})
			}
		}
	}
	return g, nil
}
