package architecture

import (
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
	"gopkg.in/yaml.v3"
)

// BuildAPIGraph extracts paths from an OpenAPI YAML file.
func BuildAPIGraph(openAPIPath string) (graph.Graph, error) {
	g := graph.Graph{Kind: "api"}
	raw, err := os.ReadFile(openAPIPath)
	if err != nil {
		if os.IsNotExist(err) {
			return g, nil
		}
		return g, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return g, err
	}
	paths, _ := doc["paths"].(map[string]any)
	for path, methods := range paths {
		pid := "path:" + path
		g.Nodes = append(g.Nodes, graph.Node{ID: pid, Kind: "path", Name: path})
		methodMap, _ := methods.(map[string]any)
		for method := range methodMap {
			if strings.HasPrefix(method, "x-") {
				continue
			}
			oid := "operation:" + strings.ToUpper(method) + " " + path
			g.Nodes = append(g.Nodes, graph.Node{ID: oid, Kind: "operation", Name: method + " " + path})
			g.Edges = append(g.Edges, graph.Edge{From: pid, To: oid, Kind: "exposes"})
		}
	}
	return g, nil
}
