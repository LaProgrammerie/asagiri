package executiongraph

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

const investigationConfidenceThreshold = 3

// ApplyInvestigationEnrichment inserts investigation nodes before risky or ambiguous work (spec §18).
func ApplyInvestigationEnrichment(nodes []GraphNode, bindings []TaskBinding, edges []GraphEdge, flow product.Flow) ([]GraphNode, []GraphEdge) {
	has := nodeIDSet(nodes)
	add := func(n GraphNode) {
		if _, ok := has[n.ID]; ok {
			return
		}
		has[n.ID] = struct{}{}
		nodes = append(nodes, n)
	}

	criticalFlow := strings.EqualFold(strings.TrimSpace(flow.Business.Criticality), "high")

	for _, b := range bindings {
		if b.NodeID == "" {
			continue
		}
		if !needsInvestigation(b, criticalFlow) {
			continue
		}
		invID := investigationNodeID(b.Action)
		if _, ok := has[invID]; ok {
			continue
		}
		reason := investigationReason(b, criticalFlow)
		add(GraphNode{
			ID:    invID,
			Type:  NodeTypeInvestigation,
			Title: fmt.Sprintf("Investigate before %s", strings.ReplaceAll(b.Action, "_", " ")),
			Agent: "local",
			Risk:  RiskLevelLow,
			Outputs: []string{
				reason,
			},
		})
		edges = append(edges, GraphEdge{
			From:   invID,
			To:     b.NodeID,
			Type:   EdgeTypeProducesContextFor,
			Reason: reason,
		})
	}

	return nodes, edges
}

func needsInvestigation(b TaskBinding, criticalFlow bool) bool {
	if strings.HasPrefix(strings.TrimSpace(b.ContractRef), "TODO:") {
		return true
	}
	if b.Sensitive {
		return true
	}
	if riskForBinding(b) == RiskLevelHigh || riskForBinding(b) == RiskLevelCritical {
		return true
	}
	if len(b.ScopePaths) >= investigationConfidenceThreshold {
		return true
	}
	if criticalFlow && b.Sensitive {
		return true
	}
	if strings.Contains(strings.ToLower(b.Action), "migration") {
		return true
	}
	return false
}

func investigationReason(b TaskBinding, criticalFlow bool) string {
	switch {
	case strings.HasPrefix(strings.TrimSpace(b.ContractRef), "TODO:"):
		return "ambiguous task: contract not finalized"
	case b.Sensitive:
		return "high-risk sensitive action requires investigation"
	case len(b.ScopePaths) >= investigationConfidenceThreshold:
		return "context too wide for direct implementation"
	case criticalFlow:
		return "critical flow impact requires investigation"
	default:
		return "low confidence task detected"
	}
}

func investigationNodeID(action string) string {
	action = strings.TrimSpace(strings.ToLower(action))
	action = strings.ReplaceAll(action, "_", "-")
	if action == "" {
		return "investigate-task"
	}
	return "investigate-" + action
}
