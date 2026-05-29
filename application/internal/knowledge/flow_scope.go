package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// FlowScopeRequest selects a flow/action slice of the knowledge graph.
type FlowScopeRequest struct {
	Flow   string
	Action string
}

// FlowScopeResult lists artefacts linked to a flow or action.
type FlowScopeResult struct {
	Flows     []string
	Files     []string
	Tests     []string
	TestFiles []string
	Symbols   []string
	APIs      []string
	Events    []string
	Metrics   []string
}

// ResolveFlowScope walks the graph from a flow and optional action.
func ResolveFlowScope(ctx context.Context, store GraphStore, req FlowScopeRequest) (FlowScopeResult, error) {
	if store == nil {
		return FlowScopeResult{}, fmt.Errorf("resolve flow scope: store required")
	}
	flow := strings.TrimSpace(req.Flow)
	if flow == "" {
		return FlowScopeResult{}, fmt.Errorf("resolve flow scope: flow required")
	}
	startID := NodeID(NodeTypeFlow, flow)
	if action := strings.TrimSpace(req.Action); action != "" {
		startID = NodeID(NodeTypeAction, action)
	}

	q := NewQuerier(store)
	nodes, _, err := q.BFS(ctx, startID, BFSOptions{MaxDepth: 12, Limit: 512})
	if err != nil {
		return FlowScopeResult{}, err
	}

	var result FlowScopeResult
	fileSet := map[string]struct{}{}
	testSet := map[string]struct{}{}
	testFileSet := map[string]struct{}{}
	symSet := map[string]struct{}{}
	apiSet := map[string]struct{}{}
	eventSet := map[string]struct{}{}
	metricSet := map[string]struct{}{}
	flowSet := map[string]struct{}{}
	if action := strings.TrimSpace(req.Action); action != "" {
		flowSet[flow+" / "+action] = struct{}{}
	} else {
		flowSet[flow] = struct{}{}
	}

	for _, n := range nodes {
		switch n.Type {
		case NodeTypeFlow:
			if n.Name != "" {
				flowSet[n.Name] = struct{}{}
			}
		case NodeTypeAction:
			if n.Name != "" && flow != "" {
				flowSet[flow+" / "+n.Name] = struct{}{}
			}
		case NodeTypeFile, NodeTypeSymbol:
			if p := strings.TrimSpace(n.Path); p != "" {
				fileSet[p] = struct{}{}
			}
			if n.Type == NodeTypeSymbol && n.Name != "" {
				symSet[n.Name] = struct{}{}
			}
		case NodeTypeTest:
			if n.Name != "" {
				testSet[n.Name] = struct{}{}
			}
			if p := strings.TrimSpace(n.Path); p != "" {
				testFileSet[p] = struct{}{}
			}
		case NodeTypeAPIOperation:
			if n.Name != "" {
				apiSet[n.Name] = struct{}{}
			}
		case NodeTypeEvent:
			if n.Name != "" {
				eventSet[n.Name] = struct{}{}
			}
		case NodeTypeMetric, NodeTypeTrace:
			if n.Name != "" {
				metricSet[n.Name] = struct{}{}
			}
		}
	}

	result.Flows = sortedKeys(flowSet)
	result.Files = sortedKeys(fileSet)
	result.Tests = sortedKeys(testSet)
	result.TestFiles = sortedKeys(testFileSet)
	result.Symbols = sortedKeys(symSet)
	result.APIs = sortedKeys(apiSet)
	result.Events = sortedKeys(eventSet)
	result.Metrics = sortedKeys(metricSet)
	return result, nil
}

// ResolveFileScope finds impact surface for a repository file path.
func ResolveFileScope(ctx context.Context, store GraphStore, filePath string) (FlowScopeResult, error) {
	if store == nil {
		return FlowScopeResult{}, fmt.Errorf("resolve file scope: store required")
	}
	filePath = strings.TrimSpace(filepathSlash(filePath))
	if filePath == "" {
		return FlowScopeResult{}, fmt.Errorf("resolve file scope: file required")
	}

	nodes, err := store.ListNodes(ctx, NodeFilter{PathLike: filePath})
	if err != nil {
		return FlowScopeResult{}, err
	}
	if len(nodes) == 0 {
		nodes, err = store.ListNodes(ctx, NodeFilter{PathLike: "%" + filepathBase(filePath)})
		if err != nil {
			return FlowScopeResult{}, err
		}
	}
	if len(nodes) == 0 {
		return FlowScopeResult{}, fmt.Errorf("%w: no graph nodes for file %q", ErrNotFound, filePath)
	}

	seen := map[string]GraphNode{}
	var edges []GraphEdge
	seenEdges := map[string]struct{}{}
	for _, seed := range nodes {
		visited, walked, err := collectReachable(ctx, store, seed.ID, 10, 256)
		if err != nil {
			return FlowScopeResult{}, err
		}
		for _, n := range visited {
			seen[n.ID] = n
		}
		for _, e := range walked {
			if _, ok := seenEdges[e.ID]; ok {
				continue
			}
			seenEdges[e.ID] = struct{}{}
			edges = append(edges, e)
		}
	}
	merged := make([]GraphNode, 0, len(seen))
	for _, n := range seen {
		merged = append(merged, n)
	}
	return scopeFromNodesWithEdges(merged, edges), nil
}

func collectReachable(ctx context.Context, store GraphStore, startID string, maxDepth, limit int) ([]GraphNode, []GraphEdge, error) {
	start, err := store.GetNode(ctx, startID)
	if err != nil {
		return nil, nil, err
	}
	if maxDepth <= 0 {
		maxDepth = 10
	}
	if limit <= 0 {
		limit = 256
	}

	visited := map[string]struct{}{startID: {}}
	var nodes []GraphNode
	var edges []GraphEdge
	seenEdges := map[string]struct{}{}
	nodes = append(nodes, start)

	type frontier struct {
		id    string
		depth int
	}
	queue := []frontier{{id: startID, depth: 0}}

	for len(queue) > 0 && len(nodes) < limit {
		cur := queue[0]
		queue = queue[1:]
		if cur.depth >= maxDepth {
			continue
		}
		incident, err := listIncidentEdges(ctx, store, cur.id)
		if err != nil {
			return nil, nil, err
		}
		for _, e := range incident {
			if _, ok := seenEdges[e.ID]; !ok {
				seenEdges[e.ID] = struct{}{}
				edges = append(edges, e)
			}
			next := e.To
			if next == cur.id {
				next = e.From
			}
			if _, seen := visited[next]; seen {
				continue
			}
			visited[next] = struct{}{}
			node, err := store.GetNode(ctx, next)
			if err != nil {
				return nil, nil, err
			}
			nodes = append(nodes, node)
			queue = append(queue, frontier{id: next, depth: cur.depth + 1})
			if len(nodes) >= limit {
				break
			}
		}
	}
	return nodes, edges, nil
}

func listIncidentEdges(ctx context.Context, store GraphStore, nodeID string) ([]GraphEdge, error) {
	out, err := store.ListEdges(ctx, EdgeFilter{FromNodeID: nodeID})
	if err != nil {
		return nil, err
	}
	in, err := store.ListEdges(ctx, EdgeFilter{ToNodeID: nodeID})
	if err != nil {
		return nil, err
	}
	return append(out, in...), nil
}

func scopeFromNodes(nodes []GraphNode) FlowScopeResult {
	return scopeFromNodesWithEdges(nodes, nil)
}

func scopeFromNodesWithEdges(nodes []GraphNode, edges []GraphEdge) FlowScopeResult {
	var result FlowScopeResult
	fileSet := map[string]struct{}{}
	testSet := map[string]struct{}{}
	testFileSet := map[string]struct{}{}
	symSet := map[string]struct{}{}
	apiSet := map[string]struct{}{}
	eventSet := map[string]struct{}{}
	flowNames := map[string]string{}
	actionNames := map[string]string{}

	for _, n := range nodes {
		switch n.Type {
		case NodeTypeFlow:
			if n.Name != "" {
				flowNames[n.ID] = n.Name
			}
		case NodeTypeAction:
			if n.Name != "" {
				actionNames[n.ID] = n.Name
			}
		case NodeTypeFile, NodeTypeSymbol:
			if p := strings.TrimSpace(n.Path); p != "" {
				fileSet[p] = struct{}{}
			}
			if n.Type == NodeTypeSymbol && n.Name != "" {
				symSet[n.Name] = struct{}{}
			}
		case NodeTypeTest:
			if n.Name != "" {
				testSet[n.Name] = struct{}{}
			}
			if p := strings.TrimSpace(n.Path); p != "" {
				testFileSet[p] = struct{}{}
			}
		case NodeTypeAPIOperation:
			if n.Name != "" {
				apiSet[n.Name] = struct{}{}
			}
		case NodeTypeEvent:
			if n.Name != "" {
				eventSet[n.Name] = struct{}{}
			}
		}
	}

	flowLineSet := map[string]struct{}{}
	for _, e := range edges {
		if e.Type != EdgeTypeRequires {
			continue
		}
		flowName, okFlow := flowNames[e.From]
		actionName, okAction := actionNames[e.To]
		if okFlow && okAction {
			flowLineSet[flowName+" / "+actionName] = struct{}{}
		}
	}
	if len(flowLineSet) == 0 {
		for _, name := range flowNames {
			flowLineSet[name] = struct{}{}
		}
	}

	result.Flows = sortedKeys(flowLineSet)
	result.Files = sortedKeys(fileSet)
	result.Tests = sortedKeys(testSet)
	result.TestFiles = sortedKeys(testFileSet)
	result.Symbols = sortedKeys(symSet)
	result.APIs = sortedKeys(apiSet)
	result.Events = sortedKeys(eventSet)
	return result
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func filepathSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func filepathBase(p string) string {
	p = filepathSlash(p)
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}
