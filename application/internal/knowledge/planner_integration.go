package knowledge

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// PlannerKnowledgeHints carries execution-graph enrichment from the knowledge graph (spec-my-E §18).
type PlannerKnowledgeHints struct {
	ValidationNodeIDs []string
	RollbackNotes     []string
	MaxParallel       int
	SchedulingWarnings []string
}

// EnrichPlannerFromGraph infers dependencies, validations, rollback, and safe parallelism (spec-my-E §18).
func EnrichPlannerFromGraph(ctx context.Context, repoRoot, flowID string, nodeOutputs map[string][]string, currentMaxParallel int) (PlannerKnowledgeHints, bool, error) {
	store, err := OpenStoreIfExists(repoRoot)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return PlannerKnowledgeHints{}, false, nil
		}
		return PlannerKnowledgeHints{}, false, err
	}
	defer store.Close()

	flowID = strings.TrimSpace(flowID)
	if flowID == "" {
		return PlannerKnowledgeHints{}, false, nil
	}

	scope, err := ResolveFlowScope(ctx, store, FlowScopeRequest{Flow: flowID})
	if err != nil {
		return PlannerKnowledgeHints{}, false, err
	}
	impact := plannerImpactFromScope(scope)

	hints := PlannerKnowledgeHints{MaxParallel: currentMaxParallel}
	if hints.MaxParallel < 1 {
		hints.MaxParallel = 2
	}

	for _, api := range scope.APIs {
		if apiHasTestLink(ctx, store, api) {
			continue
		}
		hints.ValidationNodeIDs = append(hints.ValidationNodeIDs, "validate:"+sanitizeStableKey(api))
	}

	if impact.Risk == "high" || len(impact.ImpactedFlows) > 1 {
		hints.RollbackNotes = append(hints.RollbackNotes,
			fmt.Sprintf("knowledge graph: rollback should cover %d impacted flow(s)", len(impact.ImpactedFlows)))
	}
	for _, check := range impact.RecommendedChecks {
		hints.RollbackNotes = append(hints.RollbackNotes, "recommended: "+check)
	}

	if overlap := detectParallelFileOverlap(nodeOutputs); len(overlap) > 0 {
		if hints.MaxParallel > 1 {
			hints.MaxParallel = 1
		}
		hints.SchedulingWarnings = append(hints.SchedulingWarnings,
			"knowledge graph: reduced parallelism due to shared file paths: "+strings.Join(overlap, ", "))
	}

	sort.Strings(hints.ValidationNodeIDs)
	return hints, true, nil
}

func apiHasTestLink(ctx context.Context, store GraphStore, apiName string) bool {
	nodes, err := store.ListNodes(ctx, NodeFilter{Type: NodeTypeAPIOperation})
	if err != nil {
		return false
	}
	var matched *GraphNode
	for i := range nodes {
		if nodes[i].Name == apiName || strings.Contains(nodes[i].Name, apiName) {
			matched = &nodes[i]
			break
		}
	}
	if matched == nil {
		return false
	}
	edges, err := store.ListEdges(ctx, EdgeFilter{FromNodeID: matched.ID})
	if err != nil {
		return false
	}
	for _, e := range edges {
		if e.Type == EdgeTypeTests || e.Type == EdgeTypeValidates {
			return true
		}
	}
	out, err := store.ListEdges(ctx, EdgeFilter{ToNodeID: matched.ID})
	if err != nil {
		return false
	}
	for _, e := range out {
		if e.Type == EdgeTypeTests {
			return true
		}
	}
	return false
}

func detectParallelFileOverlap(nodeOutputs map[string][]string) []string {
	fileNodes := map[string][]string{}
	for nodeID, files := range nodeOutputs {
		for _, f := range files {
			f = strings.TrimSpace(f)
			if f == "" {
				continue
			}
			fileNodes[f] = append(fileNodes[f], nodeID)
		}
	}
	var overlap []string
	for path, nodes := range fileNodes {
		if len(nodes) < 2 {
			continue
		}
		overlap = append(overlap, path)
	}
	sort.Strings(overlap)
	return overlap
}

func plannerImpactFromScope(scope FlowScopeResult) ImpactResult {
	risk := "low"
	if len(scope.APIs) >= 3 {
		risk = "medium"
	}
	if len(scope.APIs) >= 5 || (len(scope.Files) > 0 && len(scope.Tests) == 0 && len(scope.TestFiles) == 0) {
		risk = "high"
	}
	result := ImpactResult{
		ImpactedFlows: scope.Flows,
		ImpactedAPIs:  scope.APIs,
		ImpactedEvents: scope.Events,
		ImpactedTests: append(append([]string{}, scope.Tests...), scope.TestFiles...),
		Risk:          risk,
	}
	for _, t := range result.ImpactedTests {
		result.RecommendedChecks = append(result.RecommendedChecks, "go test ./... -run "+t)
	}
	return result
}

func sanitizeStableKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	return s
}
