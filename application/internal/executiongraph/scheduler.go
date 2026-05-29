package executiongraph

import (
	"context"
	"fmt"
	"sort"
)

const schedulerDefaultMaxParallel = 2

var schedulingEdgeTypes = map[EdgeType]bool{
	EdgeTypeRequires:              true,
	EdgeTypeMustRunAfter:          true,
	EdgeTypeBlocks:                true,
	EdgeTypeRollbackDependsOn:     true,
	EdgeTypeProducesContextFor:    true,
	EdgeTypeValidates:             true,
	EdgeTypeRequiresHumanApproval: true,
}

// DefaultScheduler builds parallel execution groups from graph dependencies (spec §11).
type DefaultScheduler struct{}

func (DefaultScheduler) Schedule(_ context.Context, req ScheduleRequest) (ExecutionSchedule, error) {
	graph := req.Graph
	if err := graph.Validate(); err != nil {
		return ExecutionSchedule{}, err
	}
	if err := DetectCycles(graph.Nodes, graph.Edges); err != nil {
		return ExecutionSchedule{}, err
	}

	maxParallel := graph.Strategy.MaxParallel
	if maxParallel < 1 {
		maxParallel = schedulerDefaultMaxParallel
	}
	if req.CIMode {
		maxParallel = 1
	}

	nodeIDs := sortedNodeIDs(graph.Nodes)
	inDegree, successors := buildSchedulingGraph(nodeIDs, graph.Edges)
	groups, err := scheduleParallelGroups(nodeIDs, inDegree, successors, maxParallel)
	if err != nil {
		return ExecutionSchedule{}, err
	}

	return ExecutionSchedule{
		GraphID:        graph.ID,
		ParallelGroups: groups,
		Blocked:        buildBlockedNodes(groups, graph.Edges),
	}, nil
}

func sortedNodeIDs(nodes []GraphNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}
	sort.Strings(ids)
	return ids
}

func buildSchedulingGraph(nodeIDs []string, edges []GraphEdge) (map[string]int, map[string][]string) {
	inDegree := make(map[string]int, len(nodeIDs))
	successors := make(map[string][]string, len(nodeIDs))
	for _, id := range nodeIDs {
		inDegree[id] = 0
	}

	for _, e := range edges {
		if !schedulingEdgeTypes[e.Type] {
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
	return inDegree, successors
}

func scheduleParallelGroups(
	nodeIDs []string,
	inDegree map[string]int,
	successors map[string][]string,
	maxParallel int,
) ([][]string, error) {
	indeg := make(map[string]int, len(inDegree))
	for id, d := range inDegree {
		indeg[id] = d
	}

	scheduled := make(map[string]bool, len(nodeIDs))
	groups := make([][]string, 0, len(nodeIDs))

	for len(scheduled) < len(nodeIDs) {
		ready := readyNodes(nodeIDs, scheduled, indeg)
		if len(ready) == 0 {
			return nil, fmt.Errorf("%w: scheduling stalled", ErrCycleDetected)
		}

		for len(ready) > 0 {
			batchSize := maxParallel
			if len(ready) < batchSize {
				batchSize = len(ready)
			}
			batch := append([]string(nil), ready[:batchSize]...)
			ready = ready[batchSize:]

			groups = append(groups, batch)
			for _, id := range batch {
				scheduled[id] = true
				for _, succ := range successors[id] {
					indeg[succ]--
				}
			}
		}
	}

	return groups, nil
}

func readyNodes(nodeIDs []string, scheduled map[string]bool, inDegree map[string]int) []string {
	ready := make([]string, 0)
	for _, id := range nodeIDs {
		if scheduled[id] {
			continue
		}
		if inDegree[id] == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	return ready
}

func buildBlockedNodes(groups [][]string, edges []GraphEdge) []BlockedNode {
	if len(groups) <= 1 {
		return nil
	}

	groupOf := make(map[string]int, len(groups)*2)
	for i, g := range groups {
		for _, id := range g {
			groupOf[id] = i
		}
	}

	blocked := make([]BlockedNode, 0)
	for i := 1; i < len(groups); i++ {
		for _, id := range groups[i] {
			waitFor := immediateWaitFor(schedulingPredecessors(edges, id), groupOf, i)
			if len(waitFor) == 0 {
				continue
			}
			blocked = append(blocked, BlockedNode{
				NodeID:  id,
				WaitFor: waitFor,
			})
		}
	}
	sort.Slice(blocked, func(i, j int) bool {
		return blocked[i].NodeID < blocked[j].NodeID
	})
	return blocked
}

func immediateWaitFor(preds []string, groupOf map[string]int, groupIndex int) []string {
	if len(preds) == 0 {
		return nil
	}
	prevGroup := groupIndex - 1
	waitFor := make([]string, 0, len(preds))
	for _, p := range preds {
		if groupOf[p] == prevGroup {
			waitFor = append(waitFor, p)
		}
	}
	if len(waitFor) > 0 {
		sort.Strings(waitFor)
		return waitFor
	}
	for _, p := range preds {
		if groupOf[p] < groupIndex {
			waitFor = append(waitFor, p)
		}
	}
	sort.Strings(waitFor)
	return waitFor
}

func schedulingPredecessors(edges []GraphEdge, nodeID string) []string {
	preds := make([]string, 0)
	for _, e := range edges {
		if e.To != nodeID || !schedulingEdgeTypes[e.Type] {
			continue
		}
		preds = append(preds, e.From)
	}
	sort.Strings(preds)
	return preds
}
