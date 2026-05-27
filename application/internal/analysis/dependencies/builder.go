package dependencies

import (
	"path/filepath"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/ast"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
)

// Build scans Go files for import dependencies.
func Build(repoRoot string, files []string) (graph.Graph, error) {
	g := graph.Graph{Kind: "dependency"}
	seen := map[string]struct{}{}
	for _, f := range files {
		fid := "file:" + f
		if _, ok := seen[fid]; !ok {
			seen[fid] = struct{}{}
			g.Nodes = append(g.Nodes, graph.Node{ID: fid, Kind: "file", Name: f})
		}
		parsed, err := ast.ParseGoFile(filepath.Join(repoRoot, f))
		if err != nil {
			continue
		}
		for _, imp := range parsed.Imports {
			iid := "import:" + imp
			if _, ok := seen[iid]; !ok {
				seen[iid] = struct{}{}
				g.Nodes = append(g.Nodes, graph.Node{ID: iid, Kind: "import", Name: imp})
			}
			g.Edges = append(g.Edges, graph.Edge{From: fid, To: iid, Kind: "depends_on"})
		}
	}
	return g, nil
}
