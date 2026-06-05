package knowledge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

)

const graphJSONName = "graph.json"

// BuildRequest groups inputs for graph construction (spec-my-E §22).
type BuildRequest struct {
	RepoRoot         string
	Incremental      bool
	Scope            string
	IncludeFlows     bool
	IncludeContracts bool
	IncludeCode      bool
	IncludeTests     bool
	IncludeInfra     bool
	IncludeADR       bool
	IncludeRuntime   bool
}

// BuildResult summarizes a build operation.
type BuildResult struct {
	Nodes             int      `json:"nodes"`
	Edges             int      `json:"edges"`
	Warnings          []string `json:"warnings,omitempty"`
	Rebuilt           bool     `json:"rebuilt"`
	Sources           []string `json:"sources,omitempty"`
	AvgConfidence     float64  `json:"avg_confidence,omitempty"`
	StaleFiles        int      `json:"stale_files,omitempty"`
	SkippedExtractors []string `json:"skipped_extractors,omitempty"`
}

// KnowledgeGraphBuilder constructs or updates the knowledge graph.
type KnowledgeGraphBuilder interface {
	Build(ctx context.Context, req BuildRequest) (BuildResult, error)
}

// GraphBuilder implements KnowledgeGraphBuilder using extractors and GraphStore.
type GraphBuilder struct {
	OpenStore func(string) (GraphStore, error)
}

// DefaultBuilder returns a builder using the registered SQLite store.
func DefaultBuilder() *GraphBuilder {
	return &GraphBuilder{OpenStore: OpenStore}
}

// Build extracts, upserts, and exports the knowledge graph.
func (b *GraphBuilder) Build(ctx context.Context, req BuildRequest) (BuildResult, error) {
	if req.RepoRoot == "" {
		return BuildResult{}, fmt.Errorf("knowledge build: repo root required")
	}
	open := b.OpenStore
	if open == nil {
		open = OpenStore
	}
	store, err := open(req.RepoRoot)
	if err != nil {
		return BuildResult{}, err
	}
	defer func() { _ = store.Close() }()

	rebuilt := true
	var priorMeta map[string]any
	if req.Incremental {
		meta, err := store.GetIndexMetadata(ctx, "build")
		if err == nil {
			rebuilt = false
			priorMeta = meta
		} else if !errors.Is(err, ErrNotFound) {
			return BuildResult{}, err
		}
	}

	products, err := resolveProducts(req.RepoRoot, req.Scope)
	if err != nil {
		return BuildResult{}, err
	}

	includeFlows := req.IncludeFlows
	includeContracts := req.IncludeContracts
	if !includeFlows && !includeContracts && !req.IncludeCode && !req.IncludeTests &&
		!req.IncludeInfra && !req.IncludeADR && !req.IncludeRuntime {
		includeFlows = true
		includeContracts = true
	}

	var allNodes []GraphNode
	var allEdges []GraphEdge
	var warnings []string

	categories := map[string]bool{}
	if includeFlows {
		categories["flows"] = true
	}
	if includeContracts {
		categories["contracts"] = true
	}
	if req.IncludeCode {
		categories["code"] = true
	}
	if req.IncludeTests {
		categories["tests"] = true
	}

	sourceIdx, err := ScanSourceIndex(req.RepoRoot, products, req)
	if err != nil {
		return BuildResult{}, err
	}
	storedMTimes := storedSourceMTimes(priorMeta)
	skipCats := map[string]bool{}
	if req.Incremental && priorMeta != nil {
		for cat, unix := range sourceIdx.CategoryMTimes {
			if categoryUnchanged(storedMTimes, cat, unix) {
				skipCats[string(cat)] = true
			}
		}
	}

	repoOnce := &repoOnceState{}
	for _, product := range products {
		n, e, w, err := runExtractors(ctx, req.RepoRoot, product, categories, repoOnce, skipCats)
		if err != nil {
			return BuildResult{}, err
		}
		allNodes = append(allNodes, n...)
		allEdges = append(allEdges, e...)
		warnings = append(warnings, w...)
	}

	repoCategories := map[string]bool{}
	if req.IncludeInfra {
		repoCategories["infra"] = true
	}
	if req.IncludeADR {
		repoCategories["adr"] = true
	}
	if req.IncludeRuntime {
		repoCategories["runtime"] = true
	}
	if len(repoCategories) > 0 {
		n, e, w, err := runRepoExtractors(ctx, req.RepoRoot, repoCategories, skipCats)
		if err != nil {
			return BuildResult{}, err
		}
		allNodes = append(allNodes, n...)
		allEdges = append(allEdges, e...)
		warnings = append(warnings, w...)
	}

	// Artefact extractors (spec-my-E §10): specs, tasks, trust, investigation, config.
	if includeFlows || includeContracts || req.IncludeCode || req.IncludeTests {
		artefactCats := map[string]bool{
			"specs": true, "tasks": true, "trust": true, "investigation": true, "config": true,
		}
		n, e, w, err := runRepoExtractors(ctx, req.RepoRoot, artefactCats, skipCats)
		if err != nil {
			return BuildResult{}, err
		}
		allNodes = append(allNodes, n...)
		allEdges = append(allEdges, e...)
		warnings = append(warnings, w...)
	}

	if includeFlows && req.IncludeCode {
		var linkWarn []string
		allNodes, allEdges, linkWarn = linkFlowToCode(allNodes, allEdges)
		warnings = append(warnings, linkWarn...)
	}
	if includeFlows && req.IncludeTests {
		warnings = append(warnings, warnUntestedActions(allNodes, allEdges)...)
	}

	baseNodes, baseEdges := allNodes, allEdges
	if req.Incremental && priorMeta != nil && len(skipCats) > 0 {
		existing, loadErr := store.LoadGraph(ctx)
		if loadErr == nil {
			baseNodes = append(existing.Nodes, allNodes...)
			baseEdges = append(existing.Edges, allEdges...)
		}
	}
	graph := mergeGraph(baseNodes, baseEdges)
	graph, orphanEdges := PruneOrphanEdges(graph)
	for _, edgeID := range orphanEdges {
		warnings = append(warnings, "dropped orphan edge "+edgeID)
	}
	if err := graph.Validate(); err != nil {
		return BuildResult{}, fmt.Errorf("knowledge build: %w", err)
	}

	for _, node := range graph.Nodes {
		if err := store.UpsertNode(ctx, node); err != nil {
			return BuildResult{}, err
		}
	}
	for _, edge := range graph.Edges {
		if err := store.UpsertEdge(ctx, edge); err != nil {
			return BuildResult{}, err
		}
	}

	if err := writeGraphJSON(req.RepoRoot, graph); err != nil {
		return BuildResult{}, err
	}

	buildMeta := map[string]any{
		"nodes":          len(graph.Nodes),
		"edges":          len(graph.Edges),
		"built_at":       time.Now().UTC().Format(time.RFC3339),
		"scope":          req.Scope,
		"rebuilt":        rebuilt,
		"source_mtimes":  buildSourceMTimesMap(sourceIdx),
		"include_flows":     includeFlows,
		"include_contracts": includeContracts,
		"include_code":      req.IncludeCode,
		"include_tests":     req.IncludeTests,
		"include_infra":     req.IncludeInfra,
		"include_adr":       req.IncludeADR,
		"include_runtime":   req.IncludeRuntime,
	}
	if err := store.SetIndexMetadata(ctx, "build", buildMeta); err != nil {
		return BuildResult{}, err
	}

	sources := activeSourceLabels(req, includeFlows, includeContracts)
	avgConf := averageConfidence(graph)
	skipped := skippedCategoryNames(skipCats)

	return BuildResult{
		Nodes:             len(graph.Nodes),
		Edges:             len(graph.Edges),
		Warnings:          warnings,
		Rebuilt:           rebuilt,
		Sources:           sources,
		AvgConfidence:     avgConf,
		SkippedExtractors: skipped,
	}, nil
}

func activeSourceLabels(req BuildRequest, includeFlows, includeContracts bool) []string {
	var labels []string
	if includeFlows {
		labels = append(labels, "flows")
	}
	if includeContracts {
		labels = append(labels, "contracts")
	}
	if req.IncludeCode {
		labels = append(labels, "code")
	}
	if req.IncludeTests {
		labels = append(labels, "tests")
	}
	if req.IncludeInfra {
		labels = append(labels, "infra")
	}
	if req.IncludeADR {
		labels = append(labels, "adr")
	}
	if req.IncludeRuntime {
		labels = append(labels, "runtime")
	}
	return labels
}

func averageConfidence(graph KnowledgeGraph) float64 {
	if len(graph.Nodes) == 0 && len(graph.Edges) == 0 {
		return 0
	}
	var sum float64
	var n int
	for _, node := range graph.Nodes {
		sum += node.Confidence
		n++
	}
	for _, edge := range graph.Edges {
		sum += edge.Confidence
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func skippedCategoryNames(skip map[string]bool) []string {
	if len(skip) == 0 {
		return nil
	}
	order := []string{"flows", "contracts", "code", "tests", "infra", "adr", "runtime"}
	var out []string
	for _, name := range order {
		if skip[name] {
			out = append(out, name)
		}
	}
	return out
}

func mergeGraph(nodes []GraphNode, edges []GraphEdge) KnowledgeGraph {
	nodeByID := make(map[string]GraphNode, len(nodes))
	for _, n := range nodes {
		nodeByID[n.ID] = n
	}
	edgeByID := make(map[string]GraphEdge, len(edges))
	for _, e := range edges {
		edgeByID[e.ID] = e
	}
	out := KnowledgeGraph{
		Nodes: make([]GraphNode, 0, len(nodeByID)),
		Edges: make([]GraphEdge, 0, len(edgeByID)),
	}
	for _, n := range nodeByID {
		out.Nodes = append(out.Nodes, n)
	}
	for _, e := range edgeByID {
		out.Edges = append(out.Edges, e)
	}
	return out
}

func resolveProducts(repoRoot, scope string) ([]string, error) {
	product := parseScopeProduct(scope)
	if product != "" {
		return []string{product}, nil
	}
	root := filepath.Join(repoRoot, ".asagiri", "products")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("knowledge build: no products under .asagiri/products")
		}
		return nil, err
	}
	var products []string
	for _, ent := range entries {
		if ent.IsDir() {
			products = append(products, ent.Name())
		}
	}
	if len(products) == 0 {
		return nil, fmt.Errorf("knowledge build: no products found")
	}
	return products, nil
}

func parseScopeProduct(scope string) string {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return ""
	}
	if strings.HasPrefix(scope, "product:") {
		return strings.TrimPrefix(scope, "product:")
	}
	return scope
}

// StubBuilder is deprecated; use DefaultBuilder.
type StubBuilder struct{}

func (StubBuilder) Build(ctx context.Context, req BuildRequest) (BuildResult, error) {
	return DefaultBuilder().Build(ctx, req)
}
