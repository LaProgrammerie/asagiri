package symbols

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/ast"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
)

// BuildFromGoFiles constructs a symbol graph from Go source paths (repo-relative).
func BuildFromGoFiles(repoRoot string, files []string) (graph.Graph, error) {
	g := graph.Graph{Kind: "symbol"}
	seen := map[string]struct{}{}
	for _, f := range files {
		abs := filepath.Join(repoRoot, f)
		parsed, err := ast.ParseGoFile(abs)
		if err != nil {
			continue
		}
		pkgID := "pkg:" + parsed.Package + "@" + filepath.Dir(f)
		if _, ok := seen[pkgID]; !ok {
			seen[pkgID] = struct{}{}
			g.Nodes = append(g.Nodes, graph.Node{ID: pkgID, Kind: "package", Name: parsed.Package})
		}
		for _, fn := range parsed.Funcs {
			sid := fmt.Sprintf("symbol:%s.%s", parsed.Package, fn)
			if _, ok := seen[sid]; !ok {
				seen[sid] = struct{}{}
				g.Nodes = append(g.Nodes, graph.Node{ID: sid, Kind: "func", Name: fn})
			}
			g.Edges = append(g.Edges, graph.Edge{From: sid, To: pkgID, Kind: "declared_in"})
		}
		for _, typ := range parsed.Types {
			sid := fmt.Sprintf("symbol:%s.%s", parsed.Package, typ)
			if _, ok := seen[sid]; !ok {
				seen[sid] = struct{}{}
				g.Nodes = append(g.Nodes, graph.Node{ID: sid, Kind: "type", Name: typ})
			}
			g.Edges = append(g.Edges, graph.Edge{From: sid, To: pkgID, Kind: "declared_in"})
		}
	}
	return g, nil
}

// BuildFromNames maps simple symbol names (grep hits) to nodes.
func BuildFromNames(names []string) graph.Graph {
	g := graph.Graph{Kind: "symbol"}
	for _, s := range names {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		g.Nodes = append(g.Nodes, graph.Node{ID: "symbol:" + s, Kind: "symbol", Name: s})
	}
	return g
}
