package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// AgentRoutingHint suggests a coordination role from graph signals (spec-my-E §19).
type AgentRoutingHint struct {
	Role       string
	Confidence float64
	Reason     string
}

// HandoffGraphEnrichment adds graph-derived files and constraints for handoffs.
type HandoffGraphEnrichment struct {
	Files       []string
	Constraints []string
	Symbols     []string
	APIs        []string
}

// OutputComparison scores similarity of agent outputs against graph scope.
type OutputComparison struct {
	OverlapFiles   []string
	MissingInGraph []string
	Score          float64
}

// SuggestAgentRouting returns a role override when graph signals require specialists.
func SuggestAgentRouting(ctx context.Context, store GraphStore, flow, action string, nodeRisk string) (AgentRoutingHint, bool, error) {
	if store == nil || strings.TrimSpace(flow) == "" {
		return AgentRoutingHint{}, false, nil
	}
	assess, ok, err := AssessTrustFromGraph(ctx, store, flow, action)
	if err != nil || !ok {
		return AgentRoutingHint{}, false, err
	}
	if assess.SecurityImpact == "high" || assess.ContractCoverage < 0.5 {
		return AgentRoutingHint{
			Role:       "security_auditor",
			Confidence: 0.88,
			Reason:     "knowledge graph: elevated security/contract risk",
		}, true, nil
	}
	if assess.ObservabilityCoverage < 0.4 {
		return AgentRoutingHint{
			Role:       "observability_auditor",
			Confidence: 0.82,
			Reason:     "knowledge graph: missing observability linkage",
		}, true, nil
	}
	if nodeRisk == "high" || nodeRisk == "critical" {
		return AgentRoutingHint{
			Role:       "validator",
			Confidence: 0.75,
			Reason:     "knowledge graph: high-risk node with graph scope",
		}, true, nil
	}
	return AgentRoutingHint{}, false, nil
}

// EnrichHandoffFromGraph expands handoff artefacts from flow scope.
func EnrichHandoffFromGraph(ctx context.Context, store GraphStore, flow, action string) (HandoffGraphEnrichment, error) {
	if store == nil {
		return HandoffGraphEnrichment{}, fmt.Errorf("handoff graph: store required")
	}
	scope, err := ResolveFlowScope(ctx, store, FlowScopeRequest{Flow: flow, Action: action})
	if err != nil {
		return HandoffGraphEnrichment{}, err
	}
	tests := scope.TestFiles
	if len(tests) == 0 {
		tests = scope.Tests
	}
	var constraints []string
	for _, api := range scope.APIs {
		constraints = append(constraints, "respect contract: "+api)
	}
	for _, ev := range scope.Events {
		constraints = append(constraints, "preserve event: "+ev)
	}
	return HandoffGraphEnrichment{
		Files:       scope.Files,
		Constraints: constraints,
		Symbols:     scope.Symbols,
		APIs:        scope.APIs,
	}, nil
}

// CompareOutputsAgainstGraph measures how agent file outputs align with graph scope.
func CompareOutputsAgainstGraph(ctx context.Context, store GraphStore, flow, action string, outputFiles []string) (OutputComparison, error) {
	if store == nil {
		return OutputComparison{}, fmt.Errorf("compare outputs: store required")
	}
	scope, err := ResolveFlowScope(ctx, store, FlowScopeRequest{Flow: flow, Action: action})
	if err != nil {
		return OutputComparison{}, err
	}
	graphFiles := map[string]struct{}{}
	for _, f := range scope.Files {
		graphFiles[f] = struct{}{}
	}
	var overlap, missing []string
	outSet := map[string]struct{}{}
	for _, f := range outputFiles {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		outSet[f] = struct{}{}
		if _, ok := graphFiles[f]; ok {
			overlap = append(overlap, f)
		}
	}
	for f := range graphFiles {
		if _, ok := outSet[f]; !ok {
			missing = append(missing, f)
		}
	}
	sort.Strings(overlap)
	sort.Strings(missing)
	score := 0.0
	if len(scope.Files) > 0 {
		score = float64(len(overlap)) / float64(len(scope.Files))
	}
	return OutputComparison{
		OverlapFiles:   overlap,
		MissingInGraph: missing,
		Score:          score,
	}, nil
}

// DetectSharedFileConflicts returns file paths touched by more than one node id.
func DetectSharedFileConflicts(nodeFiles map[string][]string) []string {
	return detectParallelFileOverlap(nodeFiles)
}
