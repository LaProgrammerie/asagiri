package executiongraph

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// SuggestKnowledgeDeps returns prerequisite file paths inferred from knowledge graph edges.
func SuggestKnowledgeDeps(ctx context.Context, store knowledge.GraphStore, anchorNodeID string) ([]string, error) {
	if store == nil {
		return nil, fmt.Errorf("knowledge deps: store required")
	}
	anchorNodeID = strings.TrimSpace(anchorNodeID)
	if anchorNodeID == "" {
		return nil, fmt.Errorf("knowledge deps: anchor node id required")
	}

	inEdges, err := store.ListEdges(ctx, knowledge.EdgeFilter{ToNodeID: anchorNodeID})
	if err != nil {
		return nil, err
	}

	pathSet := map[string]struct{}{}
	for _, edge := range inEdges {
		switch edge.Type {
		case knowledge.EdgeTypeRequires, knowledge.EdgeTypeDependsOn, knowledge.EdgeTypeImplements:
		default:
			continue
		}
		from, err := store.GetNode(ctx, edge.From)
		if err != nil {
			continue
		}
		if p := strings.TrimSpace(from.Path); p != "" {
			pathSet[p] = struct{}{}
		}
	}

	out := make([]string, 0, len(pathSet))
	for p := range pathSet {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// SuggestKnowledgeDepsForFile resolves a file node and returns dependency paths.
func SuggestKnowledgeDepsForFile(ctx context.Context, store knowledge.GraphStore, filePath string) ([]string, error) {
	nodes, err := store.ListNodes(ctx, knowledge.NodeFilter{Path: filePath})
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	var deps []string
	seen := map[string]struct{}{}
	for _, n := range nodes {
		part, err := SuggestKnowledgeDeps(ctx, store, n.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range part {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			deps = append(deps, p)
		}
	}
	sort.Strings(deps)
	return deps, nil
}

// AppendKnowledgeDependencyEdges adds requires edges from knowledge-graph file deps to implementation nodes.
func AppendKnowledgeDependencyEdges(ctx context.Context, repoRoot, flowID string, nodes []GraphNode, edges []GraphEdge) ([]GraphEdge, error) {
	store, err := knowledge.OpenStoreIfExists(repoRoot)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	defer store.Close()

	flowID = strings.TrimSpace(flowID)
	if flowID == "" {
		return nil, nil
	}

	actions, err := store.ListNodes(ctx, knowledge.NodeFilter{Type: knowledge.NodeTypeAction})
	if err != nil {
		return nil, err
	}

	implByFile := map[string]string{}
	for _, n := range nodes {
		if n.Type != NodeTypeImplementation {
			continue
		}
		for _, f := range n.Outputs {
			if f = strings.TrimSpace(f); f != "" {
				implByFile[f] = n.ID
			}
		}
	}

	seen := edgeKeySet(edges)
	var extra []GraphEdge
	for _, action := range actions {
		if !actionBelongsToFlow(action, flowID) {
			continue
		}
		deps, err := SuggestKnowledgeDeps(ctx, store, action.ID)
		if err != nil {
			return nil, err
		}
		for _, depPath := range deps {
			implID, ok := implByFile[depPath]
			if !ok {
				continue
			}
			fileNodes, err := store.ListNodes(ctx, knowledge.NodeFilter{Path: depPath, Type: knowledge.NodeTypeFile})
			if err != nil || len(fileNodes) == 0 {
				continue
			}
			fromID := fileNodes[0].ID
			edge := GraphEdge{
				From:   fromID,
				To:     implID,
				Type:   EdgeTypeRequires,
				Reason: "knowledge graph dependency",
			}
			key := edge.From + "|" + edge.To + "|" + string(edge.Type)
			if seen[key] {
				continue
			}
			seen[key] = true
			extra = append(extra, edge)
		}
	}
	return extra, nil
}

func actionBelongsToFlow(action knowledge.GraphNode, flowID string) bool {
	hay := strings.ToLower(action.Path + " " + action.Name)
	return strings.Contains(hay, strings.ToLower(flowID))
}

func edgeKeySet(edges []GraphEdge) map[string]bool {
	out := make(map[string]bool, len(edges))
	for _, e := range edges {
		out[e.From+"|"+e.To+"|"+string(e.Type)] = true
	}
	return out
}

// PlannerConflictsFromGraph returns scheduling warnings from knowledge-graph enrichment (spec-my-E §18).
func PlannerConflictsFromGraph(ctx context.Context, repoRoot string, graph ExecutionGraph) ([]string, error) {
	if strings.TrimSpace(repoRoot) == "" || strings.TrimSpace(graph.Flow) == "" {
		return nil, nil
	}
	nodeOutputs := make(map[string][]string, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeOutputs[n.ID] = append([]string(nil), n.Outputs...)
	}
	hints, ok, err := knowledge.EnrichPlannerFromGraph(ctx, repoRoot, graph.Flow, nodeOutputs, graph.Strategy.MaxParallel)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return hints.SchedulingWarnings, nil
}

// InferSafeParallelism caps parallelism using knowledge-graph file overlap detection (spec-my-E §18).
func InferSafeParallelism(ctx context.Context, repoRoot string, graph ExecutionGraph) (int, bool, error) {
	if strings.TrimSpace(repoRoot) == "" || strings.TrimSpace(graph.Flow) == "" {
		return graph.Strategy.MaxParallel, false, nil
	}
	nodeOutputs := make(map[string][]string, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeOutputs[n.ID] = append([]string(nil), n.Outputs...)
	}
	hints, ok, err := knowledge.EnrichPlannerFromGraph(ctx, repoRoot, graph.Flow, nodeOutputs, graph.Strategy.MaxParallel)
	if err != nil {
		return graph.Strategy.MaxParallel, false, err
	}
	if !ok || hints.MaxParallel <= 0 {
		return graph.Strategy.MaxParallel, false, nil
	}
	return hints.MaxParallel, hints.MaxParallel < graph.Strategy.MaxParallel, nil
}
