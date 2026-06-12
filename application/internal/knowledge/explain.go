package knowledge

import (
	"context"
	"fmt"
	"strings"
)

// ExplainRequest selects flow, action, and a symbol hint for path explanation.
type ExplainRequest struct {
	Flow   string
	Action string
	Symbol string
}

// ExplainStep is one hop on the shortest explanation path.
type ExplainStep struct {
	Node GraphNode  `json:"node"`
	Edge *GraphEdge `json:"edge,omitempty"`
}

// ExplainResult holds the shortest path between flow/action and a symbol.
type ExplainResult struct {
	Flow   string        `json:"flow"`
	Action string        `json:"action"`
	Symbol string        `json:"symbol"`
	Steps  []ExplainStep `json:"steps"`
}

// ExplainShortestPath finds a shortest path from action to symbol within the graph (BFS).
func (q *Querier) ExplainShortestPath(ctx context.Context, req ExplainRequest) (ExplainResult, error) {
	if q.Store == nil {
		return ExplainResult{}, fmt.Errorf("explain: store required")
	}
	flow := strings.TrimSpace(req.Flow)
	action := strings.TrimSpace(req.Action)
	symbol := strings.TrimSpace(req.Symbol)
	if flow == "" || action == "" || symbol == "" {
		return ExplainResult{}, fmt.Errorf("explain: flow, action, and symbol are required")
	}

	flowID := NodeID(NodeTypeFlow, flow)
	actionID := NodeID(NodeTypeAction, action)
	if err := ValidateNodeID(flowID); err != nil {
		return ExplainResult{}, err
	}
	if err := ValidateNodeID(actionID); err != nil {
		return ExplainResult{}, err
	}

	if _, err := q.Store.GetNode(ctx, flowID); err != nil {
		return ExplainResult{}, fmt.Errorf("explain: flow %q: %w", flow, err)
	}
	if _, err := q.Store.GetNode(ctx, actionID); err != nil {
		return ExplainResult{}, fmt.Errorf("explain: action %q: %w", action, err)
	}

	flowEdges, err := q.Store.ListEdges(ctx, EdgeFilter{
		FromNodeID: flowID,
		ToNodeID:   actionID,
		Type:       EdgeTypeRequires,
	})
	if err != nil {
		return ExplainResult{}, err
	}
	if len(flowEdges) == 0 {
		return ExplainResult{}, fmt.Errorf("explain: flow %q does not require action %q", flow, action)
	}

	goalID, err := q.resolveSymbolNode(ctx, symbol)
	if err != nil {
		return ExplainResult{}, err
	}

	steps, err := q.shortestPath(ctx, actionID, goalID)
	if err != nil {
		return ExplainResult{}, err
	}
	return ExplainResult{
		Flow:   flow,
		Action: action,
		Symbol: symbol,
		Steps:  steps,
	}, nil
}

func (q *Querier) resolveSymbolNode(ctx context.Context, hint string) (string, error) {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return "", fmt.Errorf("explain: symbol hint required")
	}
	if err := ValidateNodeID(hint); err == nil {
		if _, err := q.Store.GetNode(ctx, hint); err == nil {
			return hint, nil
		}
	}

	nodes, err := q.Store.ListNodes(ctx, NodeFilter{Type: NodeTypeSymbol})
	if err != nil {
		return "", err
	}
	lower := strings.ToLower(hint)
	var matches []GraphNode
	for _, n := range nodes {
		name := strings.ToLower(n.Name)
		id := strings.ToLower(n.ID)
		if strings.Contains(name, lower) || strings.Contains(id, lower) {
			matches = append(matches, n)
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("%w: symbol matching %q", ErrNotFound, hint)
	}
	best := matches[0]
	for _, n := range matches[1:] {
		if symbolMatchScore(n, hint) > symbolMatchScore(best, hint) {
			best = n
		}
	}
	return best.ID, nil
}

func symbolMatchScore(n GraphNode, hint string) int {
	score := 0
	lower := strings.ToLower(hint)
	if strings.EqualFold(n.Name, hint) {
		score += 100
	}
	if strings.Contains(strings.ToLower(n.Name), lower) {
		score += 50
	}
	if strings.Contains(strings.ToLower(n.ID), lower) {
		score += 10
	}
	return score
}

type pathParent struct {
	prevID string
	edge   GraphEdge
}

func (q *Querier) shortestPath(ctx context.Context, startID, goalID string) ([]ExplainStep, error) {
	if startID == goalID {
		node, err := q.Store.GetNode(ctx, startID)
		if err != nil {
			return nil, err
		}
		return []ExplainStep{{Node: node}}, nil
	}

	parent := map[string]pathParent{}
	visited := map[string]struct{}{startID: {}}
	queue := []string{startID}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		neighbors, err := q.graphNeighbors(ctx, cur)
		if err != nil {
			return nil, err
		}
		for _, nb := range neighbors {
			if _, seen := visited[nb.id]; seen {
				continue
			}
			visited[nb.id] = struct{}{}
			parent[nb.id] = pathParent{prevID: cur, edge: nb.edge}
			if nb.id == goalID {
				return q.reconstructPath(ctx, startID, goalID, parent)
			}
			queue = append(queue, nb.id)
		}
	}
	return nil, fmt.Errorf("%w: no path from %q to %q", ErrNotFound, startID, goalID)
}

type graphNeighbor struct {
	id   string
	edge GraphEdge
}

func (q *Querier) graphNeighbors(ctx context.Context, nodeID string) ([]graphNeighbor, error) {
	out, err := q.Store.ListEdges(ctx, EdgeFilter{FromNodeID: nodeID})
	if err != nil {
		return nil, err
	}
	in, err := q.Store.ListEdges(ctx, EdgeFilter{ToNodeID: nodeID})
	if err != nil {
		return nil, err
	}
	var neighbors []graphNeighbor
	for _, e := range out {
		neighbors = append(neighbors, graphNeighbor{id: e.To, edge: e})
	}
	for _, e := range in {
		neighbors = append(neighbors, graphNeighbor{
			id: e.From,
			edge: GraphEdge{
				ID:         e.ID,
				From:       e.To,
				To:         e.From,
				Type:       e.Type,
				Properties: e.Properties,
				Source:     e.Source,
				Confidence: e.Confidence,
				CreatedAt:  e.CreatedAt,
				UpdatedAt:  e.UpdatedAt,
			},
		})
	}
	return neighbors, nil
}

func (q *Querier) reconstructPath(ctx context.Context, startID, goalID string, parent map[string]pathParent) ([]ExplainStep, error) {
	var ids []string
	for at := goalID; at != startID; {
		p, ok := parent[at]
		if !ok {
			return nil, fmt.Errorf("explain: broken path reconstruction")
		}
		ids = append([]string{at}, ids...)
		at = p.prevID
	}
	ids = append([]string{startID}, ids...)

	steps := make([]ExplainStep, 0, len(ids))
	for i, id := range ids {
		node, err := q.Store.GetNode(ctx, id)
		if err != nil {
			return nil, err
		}
		step := ExplainStep{Node: node}
		if i > 0 {
			p := parent[id]
			edge := p.edge
			step.Edge = &edge
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// FormatKnowledgeExplain renders explain output for the terminal.
func FormatKnowledgeExplain(result ExplainResult) string {
	var b strings.Builder
	b.WriteString("Knowledge path\n")
	b.WriteString("──────────────\n")
	if len(result.Steps) == 0 {
		_, _ = fmt.Fprintf(&b, "No path found for %s / %s → %s\n", result.Flow, result.Action, result.Symbol)
		return b.String()
	}
	for i, step := range result.Steps {
		label := step.Node.Name
		if label == "" {
			label = step.Node.ID
		}
		if i == 0 {
			b.WriteString(label)
			continue
		}
		edgeLabel := "linked"
		if step.Edge != nil {
			edgeLabel = string(step.Edge.Type)
		}
		_, _ = fmt.Fprintf(&b, " --%s--> %s", edgeLabel, label)
	}
	b.WriteByte('\n')
	return b.String()
}
