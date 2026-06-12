package architecture

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
)

// BuildMigrationGraph links SQL migration files in order.
func BuildMigrationGraph(dirs ...string) (graph.Graph, error) {
	g := graph.Graph{Kind: "migration"}
	var files []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return g, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
				continue
			}
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	var prev string
	for _, f := range files {
		name := filepath.Base(f)
		mid := "migration:" + name
		g.Nodes = append(g.Nodes, graph.Node{ID: mid, Kind: "migration", Name: name})
		if prev != "" {
			g.Edges = append(g.Edges, graph.Edge{From: prev, To: mid, Kind: "follows"})
		}
		prev = mid
	}
	return g, nil
}
