package bus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/replay"
)

func (b *queryBus) handleGetGraphRollbackImpact(_ context.Context, q GetGraphRollbackImpactQuery) (QueryResult, error) {
	graphID := strings.TrimSpace(q.GraphID)
	if graphID == "" {
		graph, _ := latestFlowGraph(b.deps.RepoRoot, "")
		if graph != nil {
			graphID = graph.ID
		}
	}
	if graphID == "" {
		return GraphRollbackImpactResult{
			Title:          "No execution graph available to rollback.",
			ImpactLines:    []string{"Run asa plan graph or asa graph run first."},
			RollbackPolicy: "Rollback unavailable.",
			CanRollback:    false,
		}, nil
	}
	impact, err := executiongraph.AssessRollbackImpact(b.deps.RepoRoot, graphID)
	if err != nil {
		return GraphRollbackImpactResult{
			Title:          fmt.Sprintf("Rollback unavailable for %s.", graphID),
			ImpactLines:    []string{err.Error()},
			RollbackPolicy: "Fix graph state before retrying.",
			CanRollback:    false,
			CLIEquivalent:  "asa graph rollback " + graphID,
		}, nil
	}
	return GraphRollbackImpactResult{
		GraphID:        impact.GraphID,
		Title:          impact.Title,
		ImpactLines:    impact.ImpactLines,
		RollbackPolicy: impact.RollbackPolicy,
		CLIEquivalent:  impact.CLIEquivalent,
		CanRollback:    impact.RollbackPossible,
	}, nil
}

func (b *queryBus) handleGetPaletteEntries(_ context.Context, q GetPaletteEntriesQuery) (QueryResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 64
	}
	entries := append([]PaletteEntry{}, staticPaletteEntries()...)
	entries = append(entries, dynamicPaletteEntries(b.deps.RepoRoot, q.Screen)...)
	entries = appendContextualPaletteEntries(entries, b.deps.RepoRoot, q.Screen)

	query := strings.ToLower(strings.TrimSpace(q.Query))
	if query != "" {
		filtered := make([]PaletteEntry, 0, len(entries))
		for _, entry := range entries {
			if paletteEntryMatches(entry, query) {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return PaletteEntriesResult{Entries: entries}, nil
}

func staticPaletteEntries() []PaletteEntry {
	return []PaletteEntry{
		{ID: "nav.dashboard", Title: "Open dashboard", Description: "Navigate to dashboard live view", CLI: "asa dashboard", Keywords: []string{"screen", "dashboard", "nav"}, ActionID: "nav.dashboard"},
		{ID: "nav.mission", Title: "Open mission control", Description: "Navigate to mission control", CLI: "asa mission", Keywords: []string{"screen", "mission", "nav"}, ActionID: "nav.mission"},
		{ID: "nav.agents", Title: "Open agent theatre", Description: "Navigate to live agent cards", CLI: "asa agents watch", Keywords: []string{"screen", "agents", "watch", "nav"}, ActionID: "nav.agents"},
		{ID: "nav.graph", Title: "Open graph explorer", Description: "Navigate to graph explorer", CLI: "asa graph", Keywords: []string{"screen", "graph", "nav"}, ActionID: "nav.graph"},
		{ID: "nav.flow", Title: "Open flow explorer", Description: "Navigate to flow explorer", CLI: "asa flow", Keywords: []string{"screen", "flow", "nav"}, ActionID: "nav.flow"},
		{ID: "nav.logs", Title: "Open logs", Description: "Navigate to runtime logs view", CLI: "asa logs", Keywords: []string{"screen", "logs", "nav"}, ActionID: "nav.logs"},
		{ID: "nav.explain", Title: "Open explain panel", Description: "Navigate to explainability panel", CLI: "asa explain", Keywords: []string{"screen", "explain", "nav"}, ActionID: "nav.explain"},
		{ID: "nav.replay", Title: "Open replay", Description: "Navigate to replay explorer", CLI: "asa replay open <replay-id>", Keywords: []string{"screen", "replay", "nav"}, ActionID: "nav.replay"},
		{ID: "nav.prototype", Title: "Open prototype mode", Description: "Navigate to prototype split view", CLI: "asa prototype", Keywords: []string{"screen", "prototype", "nav"}, ActionID: "nav.prototype"},
		{ID: "nav.knowledge", Title: "Open knowledge", Description: "Navigate to knowledge explorer", CLI: "asa knowledge", Keywords: []string{"screen", "knowledge", "nav"}, ActionID: "nav.knowledge"},
		{ID: "nav.trust", Title: "Open trust explorer", Description: "Navigate to trust explorer", CLI: "asa trust", Keywords: []string{"screen", "trust", "nav"}, ActionID: "nav.trust"},
		{ID: "cmd.start-work", Title: "Start work", Description: "Run workflow orchestration from intent", CLI: `asa work "add workspace invitations"`, Keywords: []string{"work", "implement", "dev"}, ActionID: "cmd.start-work"},
		{ID: "cmd.run-investigation", Title: "Run investigation", Description: "Investigate onboarding failures", CLI: `asa investigate "onboarding fails"`, Keywords: []string{"investigate", "debug", "root cause"}, ActionID: "cmd.run-investigation"},
		{ID: "cmd.verify-trust", Title: "Verify trust", Description: "Run trust verification for onboarding flow", CLI: "asa verify trust onboarding", Keywords: []string{"trust", "verify", "quality"}, ActionID: "cmd.verify-trust"},
		{ID: "cmd.build-knowledge", Title: "Build knowledge graph", Description: "Rebuild the knowledge graph from repo sources", CLI: "asa knowledge build", Keywords: []string{"knowledge", "build", "graph"}, ActionID: "cmd.build-knowledge"},
		{ID: "cmd.prototype-create", Title: "Prototype create", Description: "Create product prototype from intent", CLI: `asa prototype create "<intent>"`, Keywords: []string{"prototype", "create", "product"}, ActionID: "cmd.prototype-create"},
		{ID: "cmd.flows-extract", Title: "Flows extract", Description: "Extract flows from product prototype", CLI: "asa flows extract <product>", Keywords: []string{"flows", "extract", "prototype"}, ActionID: "cmd.flows-extract"},
		{ID: "cmd.contracts-extract", Title: "Contracts extract", Description: "Extract API contracts from product flows", CLI: "asa contracts extract <product>", Keywords: []string{"contracts", "extract", "openapi"}, ActionID: "cmd.contracts-extract"},
		{ID: "cmd.spec-generate-from-product", Title: "Spec generate from product", Description: "Generate implementation spec from product artefacts", CLI: "asa spec generate-from-product <product>", Keywords: []string{"spec", "generate", "product"}, ActionID: "cmd.spec-generate-from-product"},
	}
}

func dynamicPaletteEntries(repoRoot, screen string) []PaletteEntry {
	_ = screen
	out := make([]PaletteEntry, 0, 32)

	productsRoot := filepath.Join(repoRoot, ".asagiri", "products")
	productEntries, _ := os.ReadDir(productsRoot)
	for _, productDir := range productEntries {
		if !productDir.IsDir() {
			continue
		}
		flowsDir := filepath.Join(productsRoot, productDir.Name(), "flows")
		flowFiles, err := os.ReadDir(flowsDir)
		if err != nil {
			continue
		}
		for _, flowFile := range flowFiles {
			if flowFile.IsDir() || !strings.HasSuffix(flowFile.Name(), ".flow.yaml") {
				continue
			}
			flowID := strings.TrimSuffix(flowFile.Name(), ".flow.yaml")
			out = append(out, PaletteEntry{
				ID:          "flow.open." + flowID,
				Title:       "Open flow " + flowID,
				Description: "Open flow details for " + productDir.Name(),
				CLI:         "asa flow open " + flowID,
				Keywords:    []string{"flow", flowID, productDir.Name()},
				ActionID:    "flow.open." + flowID,
			})
		}
	}

	graphsRoot := filepath.Join(repoRoot, ".asagiri", "graphs")
	graphDirs, _ := os.ReadDir(graphsRoot)
	for _, g := range graphDirs {
		if !g.IsDir() {
			continue
		}
		id := g.Name()
		out = append(out, PaletteEntry{
			ID:          "graph.status." + id,
			Title:       "Graph status " + id,
			Description: "Show execution graph status",
			CLI:         "asa graph status " + id,
			Keywords:    []string{"graph", id, "status"},
			ActionID:    "graph.status." + id,
		})
	}

	trustRoot := filepath.Join(repoRoot, ".asagiri", "trust")
	trustDirs, _ := os.ReadDir(trustRoot)
	for _, t := range trustDirs {
		if !t.IsDir() {
			continue
		}
		id := t.Name()
		out = append(out, PaletteEntry{
			ID:          "report.trust." + id,
			Title:       "Trust report " + id,
			Description: "Open trust report artefacts",
			CLI:         "asa trust show " + id,
			Keywords:    []string{"trust", "report", id},
			ActionID:    "nav.trust",
		})
	}

	replayRoot := filepath.Join(repoRoot, replay.RelDir)
	replayDirs, _ := os.ReadDir(replayRoot)
	for _, r := range replayDirs {
		if !r.IsDir() || r.Name() == replay.SnapshotsRelDir {
			continue
		}
		id := r.Name()
		out = append(out, PaletteEntry{
			ID:          "replay.open." + id,
			Title:       "Open replay " + id,
			Description: "Open replay explorer for package",
			CLI:         "asa replay open " + id,
			Keywords:    []string{"replay", id},
			ActionID:    "replay.open." + id,
		}, PaletteEntry{
			ID:          "replay.run." + id,
			Title:       "Run replay " + id,
			Description: "Execute replay package",
			CLI:         "asa replay run " + id,
			Keywords:    []string{"replay", "run", id},
			ActionID:    "replay.run." + id,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Title < out[j].Title })
	return out
}

func appendContextualPaletteEntries(entries []PaletteEntry, repoRoot, screen string) []PaletteEntry {
	switch screen {
	case "graph":
		entries = append(entries, PaletteEntry{
			ID: "ctx.graph-rollback", Title: "Rollback current graph", Description: "Destructive rollback for active graph", CLI: "asa graph rollback <graph-id>", Keywords: []string{"graph", "rollback", "destructive"}, ActionID: "safe.graph-rollback",
		})
	case "replay":
		entries = append(entries, PaletteEntry{
			ID: "ctx.replay-run", Title: "Run current replay", Description: "Execute replay for focused package", CLI: "asa replay run <replay-id>", Keywords: []string{"replay", "run"}, ActionID: "ctx.replay-run",
		})
	case "knowledge":
		entries = append(entries, PaletteEntry{
			ID: "ctx.knowledge-build", Title: "Build knowledge graph", Description: "Rebuild knowledge from sources", CLI: "asa knowledge build", Keywords: []string{"knowledge", "build"}, ActionID: "cmd.build-knowledge",
		})
	}
	if graph, warning := latestFlowGraph(repoRoot, ""); graph != nil && warning == "" {
		entries = append(entries, PaletteEntry{
			ID:          "safe.graph-rollback",
			Title:       "Rollback graph " + graph.ID,
			Description: "Destructive action requiring explicit confirmation",
			CLI:         "asa graph rollback " + graph.ID,
			Keywords:    []string{"graph", "rollback", "destructive", "safety"},
			ActionID:    "safe.graph-rollback",
		})
	}
	return entries
}

func paletteEntryMatches(entry PaletteEntry, query string) bool {
	fields := []string{
		entry.Title,
		entry.Description,
		entry.CLI,
		strings.Join(entry.Keywords, " "),
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}
