package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/architecture"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/dependencies"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/events"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/flows"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/graph"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/ownership"
	"github.com/LaProgrammerie/asagiri/application/internal/analysis/symbols"
)

const (
	productsRel  = ".asagiri/products"
	analysisRel  = ".asagiri/analysis"
	goScanRoot   = "application"
)

// Graph re-exports graph.Graph for backward compatibility.
type Graph = graph.Graph

// Node re-exports graph.Node.
type Node = graph.Node

// Edge re-exports graph.Edge.
type Edge = graph.Edge

// Bundle is the full analysis output for a product.
type Bundle = graph.Bundle

// BuildDependencyGraph scans imports for candidate files (legacy API).
func BuildDependencyGraph(repoRoot string, files []string) (Graph, error) {
	return dependencies.Build(repoRoot, files)
}

// BuildSymbolGraph maps symbols from grep hits (legacy API).
func BuildSymbolGraph(symbolsList []string) Graph {
	return symbols.BuildFromNames(symbolsList)
}

// BuildAll constructs all target graphs for a product (spec-my-A §24.16).
func BuildAll(repoRoot, productID string) (Bundle, error) {
	goFiles, err := listGoFiles(filepath.Join(repoRoot, goScanRoot))
	if err != nil {
		return Bundle{}, err
	}
	relFiles := make([]string, 0, len(goFiles))
	for _, abs := range goFiles {
		rel, err := filepath.Rel(repoRoot, abs)
		if err != nil {
			continue
		}
		relFiles = append(relFiles, filepath.ToSlash(rel))
	}

	productDir := filepath.Join(repoRoot, productsRel, productID)
	flowsDir := filepath.Join(productDir, "flows")
	contractsDir := filepath.Join(productDir, "contracts")

	b := Bundle{Product: productID, Graphs: map[string]Graph{}}

	if g, err := symbols.BuildFromGoFiles(repoRoot, relFiles); err == nil {
		b.Graphs["symbol"] = g
	}
	if g, err := dependencies.Build(repoRoot, relFiles); err == nil {
		b.Graphs["dependency"] = g
	}
	if g, err := flows.Build(flowsDir); err == nil {
		b.Graphs["flow"] = g
	}
	if g, err := events.Build(filepath.Join(contractsDir, "events.yaml")); err == nil {
		b.Graphs["event"] = g
	}
	if g, err := architecture.BuildAPIGraph(filepath.Join(contractsDir, "api.openapi.yaml")); err == nil {
		b.Graphs["api"] = g
	}
	b.Graphs["ownership"] = ownership.Build(repoRoot, relFiles)

	migDirs := []string{
		filepath.Join(repoRoot, "application/internal/runtime/migrations"),
		filepath.Join(repoRoot, "application/internal/store/sqlite/migrations"),
	}
	if g, err := architecture.BuildMigrationGraph(migDirs...); err == nil {
		b.Graphs["migration"] = g
	}

	return b, nil
}

// WriteBundle persists graphs to .asagiri/analysis/<product>/graphs.json.
func WriteBundle(repoRoot string, b Bundle) (string, error) {
	outDir := filepath.Join(repoRoot, analysisRel, b.Product)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	outPath := filepath.Join(outDir, "graphs.json")
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(outPath, raw, 0o644); err != nil {
		return "", err
	}
	return outPath, nil
}

func listGoFiles(root string) ([]string, error) {
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			out = append(out, path)
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil, nil
	}
	return out, err
}

// DefaultProduct returns the first product id under .asagiri/products when empty.
func DefaultProduct(repoRoot string) (string, error) {
	dir := filepath.Join(repoRoot, productsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("no products in %s: %w", productsRel, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			return e.Name(), nil
		}
	}
	return "", fmt.Errorf("no product directories found")
}
