package executiongraph

import "sort"

// checkpointNodeTypes defines node types that always produce a resumable checkpoint.
var checkpointNodeTypes = map[NodeType]bool{
	NodeTypeInvestigation:     true,
	NodeTypeImplementation:    true,
	NodeTypeValidation:        true,
	NodeTypeReview:            true,
	NodeTypeTrustVerification: true,
	NodeTypeManualApproval:    true,
}

// GenerateCheckpoints builds resumable checkpoints from graph topology (spec §15).
func GenerateCheckpoints(graph ExecutionGraph) []Checkpoint {
	if len(graph.Nodes) == 0 {
		return nil
	}

	order := topologicalNodeOrder(graph.Nodes, graph.Edges)
	checkpoints := make([]Checkpoint, 0, len(graph.Nodes))
	seen := make(map[string]struct{})

	for _, nodeID := range order {
		node := nodeByID(graph.Nodes, nodeID)
		if node == nil {
			continue
		}
		if !shouldCheckpoint(*node) {
			continue
		}
		if _, ok := seen[nodeID]; ok {
			continue
		}
		seen[nodeID] = struct{}{}
		checkpoints = append(checkpoints, Checkpoint{After: nodeID})
	}
	return checkpoints
}

func shouldCheckpoint(node GraphNode) bool {
	if checkpointNodeTypes[node.Type] {
		return true
	}
	return riskRank(node.Risk) >= riskRank(RiskLevelHigh)
}

func topologicalNodeOrder(nodes []GraphNode, edges []GraphEdge) []string {
	ids := sortedNodeIDs(nodes)
	inDegree := make(map[string]int, len(ids))
	successors := make(map[string][]string, len(ids))
	for _, id := range ids {
		inDegree[id] = 0
	}
	for _, e := range edges {
		if !planningEdgeTypes[e.Type] {
			continue
		}
		if _, ok := inDegree[e.From]; !ok {
			continue
		}
		if _, ok := inDegree[e.To]; !ok {
			continue
		}
		inDegree[e.To]++
		successors[e.From] = append(successors[e.From], e.To)
	}
	for from := range successors {
		sort.Strings(successors[from])
	}

	order := make([]string, 0, len(ids))
	ready := make([]string, 0)
	for _, id := range ids {
		if inDegree[id] == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)

	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]
		order = append(order, current)
		for _, succ := range successors[current] {
			inDegree[succ]--
			if inDegree[succ] == 0 {
				ready = append(ready, succ)
			}
		}
		sort.Strings(ready)
	}
	return order
}

func nodeByID(nodes []GraphNode, id string) *GraphNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}
