package knowledge

import (
	"context"
	"fmt"
)

// GraphQuery selects nodes and edges from the store (spec-my-E §6.2).
type GraphQuery struct {
	NodeID     string
	NodeType   NodeType
	Path       string
	PathPrefix string
	EdgeType   EdgeType
	FromNodeID string
	ToNodeID   string
	StartID    string
	MaxDepth   int
	Limit      int
}

// GraphQueryResult holds matched graph elements.
type GraphQueryResult struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// Querier runs graph queries and traversals.
type Querier struct {
	Store GraphStore
}

// NewQuerier creates a querier backed by store.
func NewQuerier(store GraphStore) *Querier {
	return &Querier{Store: store}
}

// Query loads nodes and edges matching filter fields.
func (q *Querier) Query(ctx context.Context, query GraphQuery) (GraphQueryResult, error) {
	if q.Store == nil {
		return GraphQueryResult{}, fmt.Errorf("query: store required")
	}
	if query.StartID != "" {
		return q.queryBFS(ctx, query)
	}
	if query.NodeID != "" {
		if err := ValidateNodeID(query.NodeID); err != nil {
			return GraphQueryResult{}, err
		}
	}
	if query.FromNodeID != "" {
		if err := ValidateNodeID(query.FromNodeID); err != nil {
			return GraphQueryResult{}, err
		}
	}
	if query.ToNodeID != "" {
		if err := ValidateNodeID(query.ToNodeID); err != nil {
			return GraphQueryResult{}, err
		}
	}

	nodes, err := q.Store.ListNodes(ctx, NodeFilter{
		ID:       query.NodeID,
		Type:     query.NodeType,
		Path:     query.Path,
		PathLike: query.PathPrefix,
	})
	if err != nil {
		return GraphQueryResult{}, err
	}

	edges, err := q.Store.ListEdges(ctx, EdgeFilter{
		Type:       query.EdgeType,
		FromNodeID: query.FromNodeID,
		ToNodeID:   query.ToNodeID,
	})
	if err != nil {
		return GraphQueryResult{}, err
	}

	if query.Limit > 0 && len(nodes) > query.Limit {
		nodes = nodes[:query.Limit]
	}
	return GraphQueryResult{Nodes: nodes, Edges: edges}, nil
}

func (q *Querier) queryBFS(ctx context.Context, query GraphQuery) (GraphQueryResult, error) {
	if err := ValidateNodeID(query.StartID); err != nil {
		return GraphQueryResult{}, err
	}
	nodes, edges, err := q.BFS(ctx, query.StartID, BFSOptions{
		MaxDepth: query.MaxDepth,
		Limit:    query.Limit,
		EdgeType: query.EdgeType,
	})
	if err != nil {
		return GraphQueryResult{}, err
	}
	return GraphQueryResult{Nodes: nodes, Edges: edges}, nil
}

// BFSOptions configures breadth-first traversal.
type BFSOptions struct {
	MaxDepth int
	Limit    int
	EdgeType EdgeType
}

// BFS walks outgoing edges from startID.
func (q *Querier) BFS(ctx context.Context, startID string, opts BFSOptions) ([]GraphNode, []GraphEdge, error) {
	if err := ValidateNodeID(startID); err != nil {
		return nil, nil, err
	}
	start, err := q.Store.GetNode(ctx, startID)
	if err != nil {
		return nil, nil, err
	}

	maxDepth := opts.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 16
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 256
	}

	visitedNodes := map[string]struct{}{startID: {}}
	visitedEdges := map[string]struct{}{}
	var nodes []GraphNode
	var edges []GraphEdge
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

		outEdges, err := q.Store.ListEdges(ctx, EdgeFilter{FromNodeID: cur.id, Type: opts.EdgeType})
		if err != nil {
			return nil, nil, err
		}
		for _, e := range outEdges {
			if _, seen := visitedEdges[e.ID]; seen {
				continue
			}
			visitedEdges[e.ID] = struct{}{}
			edges = append(edges, e)
			if len(nodes) >= limit {
				break
			}
			if _, seen := visitedNodes[e.To]; seen {
				continue
			}
			next, err := q.Store.GetNode(ctx, e.To)
			if err != nil {
				return nil, nil, err
			}
			visitedNodes[e.To] = struct{}{}
			nodes = append(nodes, next)
			queue = append(queue, frontier{id: e.To, depth: cur.depth + 1})
			if len(nodes) >= limit {
				break
			}
		}
	}

	return nodes, edges, nil
}

// BFSUndirected walks edges in both directions from startID.
func (q *Querier) BFSUndirected(ctx context.Context, startID string, opts BFSOptions) ([]GraphNode, []GraphEdge, error) {
	if err := ValidateNodeID(startID); err != nil {
		return nil, nil, err
	}
	start, err := q.Store.GetNode(ctx, startID)
	if err != nil {
		return nil, nil, err
	}

	maxDepth := opts.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 16
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 256
	}

	visitedNodes := map[string]struct{}{startID: {}}
	visitedEdges := map[string]struct{}{}
	var nodes []GraphNode
	var edges []GraphEdge
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

		outEdges, err := q.Store.ListEdges(ctx, EdgeFilter{FromNodeID: cur.id, Type: opts.EdgeType})
		if err != nil {
			return nil, nil, err
		}
		inEdges, err := q.Store.ListEdges(ctx, EdgeFilter{ToNodeID: cur.id, Type: opts.EdgeType})
		if err != nil {
			return nil, nil, err
		}
		for _, e := range append(outEdges, inEdges...) {
			if _, seen := visitedEdges[e.ID]; seen {
				continue
			}
			visitedEdges[e.ID] = struct{}{}
			edges = append(edges, e)
			nextID := e.To
			if nextID == cur.id {
				nextID = e.From
			}
			if _, seen := visitedNodes[nextID]; seen {
				continue
			}
			next, err := q.Store.GetNode(ctx, nextID)
			if err != nil {
				return nil, nil, err
			}
			visitedNodes[nextID] = struct{}{}
			nodes = append(nodes, next)
			queue = append(queue, frontier{id: nextID, depth: cur.depth + 1})
			if len(nodes) >= limit {
				break
			}
		}
	}

	return nodes, edges, nil
}
