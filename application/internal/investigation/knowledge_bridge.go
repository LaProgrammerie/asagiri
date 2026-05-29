package investigation

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// GraphScopeOptions configures knowledge-graph scope resolution.
type GraphScopeOptions struct {
	UseKnowledgeGraph bool
	Flow              string
	Action            string
}

// ResolveScopeFromGraph loads artefacts for a flow/action from the knowledge graph (spec-my-E §15–16).
func ResolveScopeFromGraph(ctx context.Context, repoRoot string, opts GraphScopeOptions) (ContextPack, error) {
	if !opts.UseKnowledgeGraph {
		return ContextPack{}, fmt.Errorf("knowledge graph scope: disabled")
	}
	store, err := knowledge.OpenStoreIfExists(repoRoot)
	if err != nil {
		return ContextPack{}, err
	}
	if store == nil {
		return ContextPack{}, knowledge.ErrNotFound
	}
	defer store.Close()

	scope, err := knowledge.ResolveFlowScope(ctx, store, knowledge.FlowScopeRequest{
		Flow:   opts.Flow,
		Action: opts.Action,
	})
	if err != nil {
		return ContextPack{}, err
	}
	return contextPackFromFlowScope(scope), nil
}

func contextPackFromFlowScope(scope knowledge.FlowScopeResult) ContextPack {
	tests := scope.TestFiles
	if len(tests) == 0 {
		tests = scope.Tests
	}
	pack := ContextPack{
		Files:     scope.Files,
		Tests:     tests,
		Symbols:   scope.Symbols,
		Contracts: scope.APIs,
		APIs:      scope.APIs,
		Events:    scope.Events,
		Metrics:   scope.Metrics,
		Flows:     scope.Flows,
		Risks:     graphRisksFromScope(scope),
	}
	if pack.MaxFiles == 0 {
		pack.MaxFiles = 80
	}
	if len(pack.Files) > pack.MaxFiles {
		pack.Files = pack.Files[:pack.MaxFiles]
	}
	return pack
}

func graphRisksFromScope(scope knowledge.FlowScopeResult) []string {
	var risks []string
	if len(scope.APIs) == 0 && len(scope.Files) > 0 {
		risks = append(risks, "code paths without linked API operations in knowledge graph")
	}
	if len(scope.Tests) == 0 && len(scope.TestFiles) == 0 && len(scope.Files) > 0 {
		risks = append(risks, "no tests linked in knowledge graph for scoped paths")
	}
	if len(scope.Events) == 0 && len(scope.APIs) > 0 {
		risks = append(risks, "API surface without linked domain events")
	}
	return risks
}

// MergeGraphScope merges graph-resolved artefacts into an investigation context pack.
func MergeGraphScope(pack ContextPack, graph ContextPack) ContextPack {
	pack.Files = uniqueStrings(append(pack.Files, graph.Files...))
	pack.Tests = uniqueStrings(append(pack.Tests, graph.Tests...))
	pack.Symbols = uniqueStrings(append(pack.Symbols, graph.Symbols...))
	pack.Contracts = uniqueStrings(append(pack.Contracts, graph.Contracts...))
	pack.APIs = uniqueStrings(append(pack.APIs, graph.APIs...))
	pack.Events = uniqueStrings(append(pack.Events, graph.Events...))
	pack.Metrics = uniqueStrings(append(pack.Metrics, graph.Metrics...))
	pack.Flows = uniqueStrings(append(pack.Flows, graph.Flows...))
	pack.Risks = uniqueStrings(append(pack.Risks, graph.Risks...))
	if pack.MaxFiles == 0 {
		pack.MaxFiles = 80
	}
	if len(pack.Files) > pack.MaxFiles {
		pack.Files = pack.Files[:pack.MaxFiles]
	}
	return pack
}

// EnrichFromKnowledgeGraph merges graph scope into investigation results when graph.sqlite exists.
func EnrichFromKnowledgeGraph(ctx context.Context, repoRoot string, scope *ResolvedScope, local *InvestigationResult) (ContextPack, bool) {
	if scope == nil || local == nil || scope.Flow == "" {
		return ContextPack{}, false
	}
	action := scope.Action
	graphPack, err := ResolveScopeFromGraph(ctx, repoRoot, GraphScopeOptions{
		UseKnowledgeGraph: true,
		Flow:              scope.Flow,
		Action:            action,
	})
	if err != nil {
		return ContextPack{}, false
	}
	local.CandidateFiles = uniqueStrings(append(local.CandidateFiles, graphPack.Files...))
	local.RelatedTests = uniqueStrings(append(local.RelatedTests, graphPack.Tests...))
	local.Symbols = uniqueStrings(append(local.Symbols, graphPack.Symbols...))
	scope.Contracts = uniqueStrings(append(scope.Contracts, graphPack.Contracts...))
	if scope.Action == "" && len(graphPack.Flows) > 0 {
		for _, line := range graphPack.Flows {
			if _, act := splitFlowAction(line); act != "" {
				scope.Action = act
				break
			}
		}
	}
	return graphPack, true
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// LinkLogsToFlowEvents matches symptom tokens to graph event names (spec-my-E §16).
func LinkLogsToFlowEvents(symptom string, events []string) []string {
	symptom = strings.ToLower(symptom)
	var hints []string
	for _, ev := range events {
		evLower := strings.ToLower(ev)
		token := strings.ReplaceAll(evLower, ".", " ")
		if strings.Contains(symptom, evLower) || strings.Contains(symptom, token) {
			hints = append(hints, "log pattern may relate to event: "+ev)
		}
	}
	return hints
}

func splitFlowAction(line string) (flow, action string) {
	line = strings.TrimSpace(line)
	if before, after, ok := strings.Cut(line, " / "); ok {
		return strings.TrimSpace(before), strings.TrimSpace(after)
	}
	return line, ""
}

// BuildRootCauseGraphWithKnowledge merges investigation report data with knowledge-graph scope (spec-my-E §16).
func BuildRootCauseGraphWithKnowledge(ctx context.Context, repoRoot string, rep Report, pack ContextPack) (RootCauseGraph, error) {
	g := BuildRootCauseGraph(rep, pack)
	if strings.TrimSpace(repoRoot) == "" || strings.TrimSpace(rep.Scope.Flow) == "" {
		return g, nil
	}
	graphPack, err := ResolveScopeFromGraph(ctx, repoRoot, GraphScopeOptions{
		UseKnowledgeGraph: true,
		Flow:              rep.Scope.Flow,
		Action:            rep.Scope.Action,
	})
	if err != nil {
		return g, err
	}
	return AugmentRootCauseGraph(g, rep, graphPack), nil
}

// AugmentRootCauseGraph adds symbols, tests, and metrics from a knowledge context pack.
func AugmentRootCauseGraph(g RootCauseGraph, rep Report, pack ContextPack) RootCauseGraph {
	seen := rootCauseNodeIDs(g.Nodes)
	addNode := func(id, kind, label string) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		g.Nodes = append(g.Nodes, GraphNode{ID: id, Kind: kind, Label: label})
	}
	flowID := ""
	if rep.Scope.Flow != "" {
		flowID = "flow:" + rep.Scope.Flow
	}
	for i, sym := range pack.Symbols {
		nid := fmt.Sprintf("symbol:%d", i)
		addNode(nid, "symbol", sym)
		if flowID != "" {
			g.Edges = append(g.Edges, GraphEdge{From: flowID, To: nid, Relation: "implements"})
		}
	}
	for i, test := range pack.Tests {
		nid := fmt.Sprintf("test:%d", i)
		addNode(nid, "test", test)
		if flowID != "" {
			g.Edges = append(g.Edges, GraphEdge{From: flowID, To: nid, Relation: "tests"})
		}
	}
	for i, metric := range pack.Metrics {
		nid := fmt.Sprintf("metric:%d", i)
		addNode(nid, "metric", metric)
		if flowID != "" {
			g.Edges = append(g.Edges, GraphEdge{From: flowID, To: nid, Relation: "observes"})
		}
	}
	for _, risk := range pack.Risks {
		rid := "risk:" + sanitizeGraphKey(risk)
		addNode(rid, "risk", risk)
		g.Edges = append(g.Edges, GraphEdge{From: "symptom:0", To: rid, Relation: "risk"})
	}
	return g
}

func rootCauseNodeIDs(nodes []GraphNode) map[string]struct{} {
	out := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		out[n.ID] = struct{}{}
	}
	return out
}

func sanitizeGraphKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	return strings.NewReplacer("/", "_", ".", "_", ":", "_").Replace(s)
}
