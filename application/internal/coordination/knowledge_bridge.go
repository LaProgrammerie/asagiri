package coordination

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// OpenKnowledgeStoreIfPresent opens the knowledge graph when graph.sqlite exists.
func OpenKnowledgeStoreIfPresent(repoRoot string) (knowledge.GraphStore, error) {
	store, err := knowledge.OpenStoreIfExists(repoRoot)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return store, nil
}

// EnrichContextFromGraph adds graph-linked artefacts to a coordination context pack (spec-my-E §15, §19).
func EnrichContextFromGraph(ctx context.Context, pack ContextPack, store knowledge.GraphStore, flow, action string) (ContextPack, error) {
	if store == nil || strings.TrimSpace(flow) == "" {
		return pack, nil
	}
	scope, err := knowledge.ResolveFlowScope(ctx, store, knowledge.FlowScopeRequest{
		Flow:   flow,
		Action: action,
	})
	if err != nil {
		return pack, err
	}

	fileSet := make(map[string]struct{}, len(pack.Files)+len(scope.Files))
	for _, f := range pack.Files {
		fileSet[f] = struct{}{}
	}
	for _, f := range scope.Files {
		fileSet[f] = struct{}{}
	}
	pack.Files = make([]string, 0, len(fileSet))
	for f := range fileSet {
		pack.Files = append(pack.Files, f)
	}
	sort.Strings(pack.Files)

	for _, api := range scope.APIs {
		pack.TrustConstraints = append(pack.TrustConstraints, "graph-api:"+api)
	}
	for _, ev := range scope.Events {
		pack.InvestigationOutputs = append(pack.InvestigationOutputs, "graph-event:"+ev)
	}
	pack.TrustConstraints = uniqueStringsCoord(pack.TrustConstraints)
	pack.InvestigationOutputs = uniqueStringsCoord(pack.InvestigationOutputs)

	if len(scope.Flows) > 0 && pack.Flow == "" {
		pack.Flow = scope.Flows[0]
	}
	return pack, nil
}

// ApplyGraphAgentRouting overrides assignment role when the knowledge graph requires a specialist (§19).
func ApplyGraphAgentRouting(ctx context.Context, store knowledge.GraphStore, flow, action string, node executiongraph.GraphNode, asg AgentAssignment) (AgentAssignment, error) {
	if store == nil || strings.TrimSpace(flow) == "" {
		return asg, nil
	}
	hint, ok, err := knowledge.SuggestAgentRouting(ctx, store, flow, action, string(node.Risk))
	if err != nil {
		return asg, err
	}
	if !ok {
		return asg, nil
	}
	role := AgentRole(hint.Role)
	if err := ValidateRole(role); err != nil {
		return asg, nil
	}
	asg.Role = role
	if asg.ProfileID == "" {
		asg.ProfileID = "knowledge-graph"
	}
	return asg, nil
}

// EnrichHandoffWithGraph merges graph constraints and files into a handoff (§19).
func EnrichHandoffWithGraph(ctx context.Context, store knowledge.GraphStore, flow, action string, h Handoff) (Handoff, error) {
	if store == nil || strings.TrimSpace(flow) == "" {
		return h, nil
	}
	enrich, err := knowledge.EnrichHandoffFromGraph(ctx, store, flow, action)
	if err != nil {
		return h, err
	}
	h.Files = uniqueStringsCoord(append(h.Files, enrich.Files...))
	h.Constraints = uniqueStringsCoord(append(h.Constraints, enrich.Constraints...))
	if h.Confidence == 0 {
		h.Confidence = 0.85
	}
	return h, nil
}

// DetectKnowledgeConflicts finds file overlaps using knowledge graph scope (§19).
func DetectKnowledgeConflicts(ctx context.Context, graph ExecutionGraph, store knowledge.GraphStore) ([]Conflict, error) {
	if store == nil {
		return nil, nil
	}
	nodeFiles := make(map[string][]string)
	for _, n := range graph.Nodes {
		nodeFiles[n.ID] = append(nodeFiles[n.ID], n.Outputs...)
	}
	shared := knowledge.DetectSharedFileConflicts(nodeFiles)
	if len(shared) == 0 {
		return nil, nil
	}
	var out []Conflict
	for _, path := range shared {
		var nodes []string
		for id, files := range nodeFiles {
			for _, f := range files {
				if f == path {
					nodes = append(nodes, id)
					break
				}
			}
		}
		out = append(out, Conflict{
			Category: ConflictFileOverlap,
			Message:  fmt.Sprintf("knowledge graph: concurrent edit risk on %q", path),
			NodeIDs:  nodes,
		})
	}
	scope, err := knowledge.ResolveFlowScope(ctx, store, knowledge.FlowScopeRequest{Flow: graph.Flow})
	if err == nil && len(scope.APIs) > 1 {
		out = append(out, Conflict{
			Category: ConflictContractDrift,
			Message:  fmt.Sprintf("knowledge graph: %d APIs in flow scope may drift together", len(scope.APIs)),
		})
	}
	return out, nil
}

// CompareAgentOutputs scores agent output files against graph scope (§19).
func CompareAgentOutputs(ctx context.Context, store knowledge.GraphStore, flow, action string, files []string) (knowledge.OutputComparison, error) {
	if store == nil {
		return knowledge.OutputComparison{}, fmt.Errorf("compare outputs: knowledge store required")
	}
	return knowledge.CompareOutputsAgainstGraph(ctx, store, flow, action, files)
}

// KnowledgeAwareConflictDetector wraps the default detector with graph-backed conflicts.
type KnowledgeAwareConflictDetector struct {
	RepoRoot string
	Inner    ConflictDetector
}

// Detect runs default detection plus knowledge graph conflicts.
func (d KnowledgeAwareConflictDetector) Detect(ctx context.Context, graph ExecutionGraph) ([]Conflict, error) {
	inner := d.Inner
	if inner == nil {
		inner = DefaultConflictDetector{}
	}
	conflicts, err := inner.Detect(ctx, graph)
	if err != nil {
		return nil, err
	}
	store, err := OpenKnowledgeStoreIfPresent(d.RepoRoot)
	if err != nil {
		return conflicts, err
	}
	if store == nil {
		return conflicts, nil
	}
	defer store.Close()
	extra, err := DetectKnowledgeConflicts(ctx, graph, store)
	if err != nil {
		return conflicts, err
	}
	return append(conflicts, extra...), nil
}

func uniqueStringsCoord(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
