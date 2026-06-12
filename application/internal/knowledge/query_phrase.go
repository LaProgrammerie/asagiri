package knowledge

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	phraseImplements = regexp.MustCompile(`(?i)^what implements\s+(.+?)\??$`)
	phraseTestsCover = regexp.MustCompile(`(?i)^which tests cover\s+(.+?)\??$`)
)

// ParsedPhrase is the structured interpretation of a natural-language query.
type ParsedPhrase struct {
	GraphQuery GraphQuery
	Label      string
}

// ParseQueryPhrase maps known question patterns to graph queries.
func ParseQueryPhrase(phrase string) (ParsedPhrase, bool) {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return ParsedPhrase{}, false
	}
	if m := phraseImplements.FindStringSubmatch(phrase); len(m) == 2 {
		name := strings.TrimSpace(m[1])
		return ParsedPhrase{
			Label: "implements:" + name,
			GraphQuery: GraphQuery{
				Limit: 64,
			},
		}, true
	}
	if m := phraseTestsCover.FindStringSubmatch(phrase); len(m) == 2 {
		flow := strings.TrimSpace(m[1])
		flowID := NodeID(NodeTypeFlow, flow)
		return ParsedPhrase{
			Label: "tests_cover:" + flow,
			GraphQuery: GraphQuery{
				StartID:  flowID,
				MaxDepth: 8,
				Limit:    64,
			},
		}, true
	}
	return ParsedPhrase{}, false
}

// QueryImplements resolves "what implements X" by walking action → API → symbol.
func (q *Querier) QueryImplements(ctx context.Context, name string) (GraphQueryResult, error) {
	actionID, err := q.findNodeIDByName(ctx, NodeTypeAction, name)
	if err != nil {
		return GraphQueryResult{}, err
	}
	action, err := q.Store.GetNode(ctx, actionID)
	if err != nil {
		return GraphQueryResult{}, err
	}

	var nodes []GraphNode
	var edges []GraphEdge
	seenNodes := map[string]struct{}{actionID: {}}
	seenEdges := map[string]struct{}{}
	nodes = append(nodes, action)

	addNode := func(n GraphNode) {
		if _, ok := seenNodes[n.ID]; ok {
			return
		}
		seenNodes[n.ID] = struct{}{}
		nodes = append(nodes, n)
	}
	addEdge := func(e GraphEdge) {
		if _, ok := seenEdges[e.ID]; ok {
			return
		}
		seenEdges[e.ID] = struct{}{}
		edges = append(edges, e)
	}

	requires, err := q.Store.ListEdges(ctx, EdgeFilter{FromNodeID: actionID, Type: EdgeTypeRequires})
	if err != nil {
		return GraphQueryResult{}, err
	}
	for _, reqEdge := range requires {
		addEdge(reqEdge)
		api, err := q.Store.GetNode(ctx, reqEdge.To)
		if err != nil {
			continue
		}
		addNode(api)
		implEdges, err := q.Store.ListEdges(ctx, EdgeFilter{FromNodeID: api.ID, Type: EdgeTypeImplements})
		if err != nil {
			return GraphQueryResult{}, err
		}
		for _, impl := range implEdges {
			addEdge(impl)
			sym, err := q.Store.GetNode(ctx, impl.To)
			if err != nil {
				continue
			}
			addNode(sym)
		}
	}
	return GraphQueryResult{Nodes: nodes, Edges: edges}, nil
}

// QueryTestsCover returns tests reachable from a flow via graph traversal.
func (q *Querier) QueryTestsCover(ctx context.Context, flowName string) (GraphQueryResult, error) {
	flowID := NodeID(NodeTypeFlow, flowName)
	if _, err := q.Store.GetNode(ctx, flowID); err != nil {
		flowID, err = q.findNodeIDByName(ctx, NodeTypeFlow, flowName)
		if err != nil {
			return GraphQueryResult{}, err
		}
	}
	nodes, edges, err := q.BFS(ctx, flowID, BFSOptions{MaxDepth: 8, Limit: 64})
	if err != nil {
		return GraphQueryResult{}, err
	}
	var tests []GraphNode
	for _, n := range nodes {
		if n.Type == NodeTypeTest {
			tests = append(tests, n)
		}
	}
	if len(tests) == 0 {
		return GraphQueryResult{Nodes: nodes, Edges: edges}, nil
	}
	return GraphQueryResult{Nodes: tests, Edges: edges}, nil
}

func (q *Querier) findNodeIDByName(ctx context.Context, nodeType NodeType, name string) (string, error) {
	name = strings.TrimSpace(name)
	nodes, err := q.Store.ListNodes(ctx, NodeFilter{Type: nodeType})
	if err != nil {
		return "", err
	}
	var matches []GraphNode
	for _, n := range nodes {
		if strings.EqualFold(n.Name, name) || strings.EqualFold(strings.TrimPrefix(n.ID, string(nodeType)+":"), name) {
			matches = append(matches, n)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%w: %s %q", ErrNotFound, nodeType, name)
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("ambiguous %s name %q (%d matches)", nodeType, name, len(matches))
	}
}

// RunPhraseQuery executes a parsed phrase query.
func (q *Querier) RunPhraseQuery(ctx context.Context, parsed ParsedPhrase) (GraphQueryResult, error) {
	if strings.HasPrefix(parsed.Label, "implements:") {
		name := strings.TrimPrefix(parsed.Label, "implements:")
		return q.QueryImplements(ctx, name)
	}
	if strings.HasPrefix(parsed.Label, "tests_cover:") {
		flow := strings.TrimPrefix(parsed.Label, "tests_cover:")
		return q.QueryTestsCover(ctx, flow)
	}
	return q.Query(ctx, parsed.GraphQuery)
}
