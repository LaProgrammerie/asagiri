package executiongraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// ApplyKnowledgeGraphEnrichment adds validation nodes, rollback notes, and parallelism hints (spec-my-E §18).
func ApplyKnowledgeGraphEnrichment(ctx context.Context, repoRoot string, graph *ExecutionGraph) error {
	if graph == nil || strings.TrimSpace(repoRoot) == "" || strings.TrimSpace(graph.Flow) == "" {
		return nil
	}
	nodeOutputs := make(map[string][]string, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeOutputs[n.ID] = append([]string(nil), n.Outputs...)
	}
	hints, ok, err := knowledge.EnrichPlannerFromGraph(ctx, repoRoot, graph.Flow, nodeOutputs, graph.Strategy.MaxParallel)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if hints.MaxParallel > 0 && hints.MaxParallel < graph.Strategy.MaxParallel {
		graph.Strategy.MaxParallel = hints.MaxParallel
	}

	for _, valID := range hints.ValidationNodeIDs {
		if nodeExists(graph.Nodes, valID) {
			continue
		}
		graph.Nodes = append(graph.Nodes, GraphNode{
			ID:     valID,
			Type:   NodeTypeValidation,
			Title:  "Knowledge graph validation: " + valID,
			Risk:   RiskLevelMedium,
			Status: NodeStatusPending,
			RequiredChecks: []string{
				"knowledge-graph",
				"tests",
			},
		})
		graph.Edges = append(graph.Edges, GraphEdge{
			From:   lastImplementationNodeID(graph.Nodes),
			To:     valID,
			Type:   EdgeTypeValidates,
			Reason: "knowledge graph: API without linked test",
		})
	}

	if len(hints.RollbackNotes) > 0 {
		if graph.Rollback == nil {
			graph.Rollback = &RollbackPlan{Strategy: RollbackStrategyWorktreeReset, PreserveReports: true}
		}
		for i := range graph.Nodes {
			if graph.Nodes[i].Type == NodeTypeImplementation && graph.Nodes[i].RollbackStrategy == "" {
				graph.Nodes[i].RollbackStrategy = RollbackStrategyPatchRevert
			}
		}
	}
	return nil
}

func nodeExists(nodes []GraphNode, id string) bool {
	for _, n := range nodes {
		if n.ID == id {
			return true
		}
	}
	return false
}

func lastImplementationNodeID(nodes []GraphNode) string {
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodes[i].Type == NodeTypeImplementation {
			return nodes[i].ID
		}
	}
	if len(nodes) > 0 {
		return nodes[0].ID
	}
	return fmt.Sprintf("node-%d", 0)
}
