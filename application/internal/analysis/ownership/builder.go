package ownership

import (
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
)

var prefixOwners = []struct {
	prefix string
	owner  string
}{
	{"application/internal/runtime/", "runtime-team"},
	{"application/internal/investigation/", "investigation-team"},
	{"application/internal/analysis/", "platform-team"},
	{"application/internal/product/", "product-team"},
	{"application/internal/cli/", "cli-team"},
	{"application/pkg/asagiri/", "sdk-team"},
	{"docs-site/", "docs-team"},
}

// Build assigns heuristic owners from file paths.
func Build(repoRoot string, files []string) graph.Graph {
	g := graph.Graph{Kind: "ownership"}
	seen := map[string]struct{}{}
	for _, f := range files {
		fid := "file:" + f
		g.Nodes = append(g.Nodes, graph.Node{ID: fid, Kind: "file", Name: f})
		owner := ownerForPath(f)
		oid := "owner:" + owner
		if _, ok := seen[oid]; !ok {
			seen[oid] = struct{}{}
			g.Nodes = append(g.Nodes, graph.Node{ID: oid, Kind: "owner", Name: owner})
		}
		g.Edges = append(g.Edges, graph.Edge{From: fid, To: oid, Kind: "owned_by"})
		// package heuristic
		if strings.HasSuffix(f, ".go") {
			pkg := filepath.ToSlash(filepath.Dir(f))
			pkgID := "package:" + pkg
			if _, ok := seen[pkgID]; !ok {
				seen[pkgID] = struct{}{}
				g.Nodes = append(g.Nodes, graph.Node{ID: pkgID, Kind: "package", Name: pkg})
				g.Edges = append(g.Edges, graph.Edge{From: pkgID, To: oid, Kind: "owned_by"})
			}
			g.Edges = append(g.Edges, graph.Edge{From: fid, To: pkgID, Kind: "in_package"})
		}
	}
	_ = repoRoot
	return g
}

func ownerForPath(rel string) string {
	rel = filepath.ToSlash(rel)
	for _, rule := range prefixOwners {
		if strings.HasPrefix(rel, rule.prefix) {
			return rule.owner
		}
	}
	if strings.HasPrefix(rel, "application/") {
		return "application-default"
	}
	return "unassigned"
}
