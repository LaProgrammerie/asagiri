package executiongraph

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

// TrustEnrichmentInput configures trust gate insertion during planning (spec §17).
type TrustEnrichmentInput struct {
	Flow                     product.Flow
	Gates                    TrustGateConfig
	TrustRequiredForHighRisk bool
}

// TrustGateConfig mirrors execution_graph.gates for planner hooks.
type TrustGateConfig struct {
	TrustRequiredForHighRisk bool
}

// ApplyTrustEnrichment inserts trust verification nodes and required checks (spec §17).
func ApplyTrustEnrichment(nodes []GraphNode, bindings []TaskBinding, edges []GraphEdge, input TrustEnrichmentInput) ([]GraphNode, []GraphEdge) {
	has := nodeIDSet(nodes)
	add := func(n GraphNode) {
		if _, ok := has[n.ID]; ok {
			return
		}
		has[n.ID] = struct{}{}
		nodes = append(nodes, n)
	}

	trustRequired := input.TrustRequiredForHighRisk || input.Gates.TrustRequiredForHighRisk
	criticalFlow := strings.EqualFold(strings.TrimSpace(input.Flow.Business.Criticality), "high")
	hasObservability := len(input.Flow.Observability.Metrics) > 0 || len(input.Flow.Observability.Traces) > 0

	for i := range nodes {
		switch nodes[i].Type {
		case NodeTypeValidation:
			if criticalFlow {
				nodes[i].RequiredChecks = appendUnique(nodes[i].RequiredChecks, "flows")
			}
			if !hasObservability {
				nodes[i].RequiredChecks = appendUnique(nodes[i].RequiredChecks, "observability")
			}
		case NodeTypeImplementation:
			binding := bindingForNode(bindings, nodes[i].ID)
			if binding != nil && isPublicContract(binding.ContractRef) {
				nodes[i].RequiredChecks = appendUnique(nodes[i].RequiredChecks, "backward_compatibility", "trust")
			}
			if binding != nil && binding.Sensitive {
				nodes[i].RequiredChecks = appendUnique(nodes[i].RequiredChecks, "security")
			}
		}
	}

	for _, b := range bindings {
		needsGate := isPublicContract(b.ContractRef) || (trustRequired && (b.Sensitive || riskForBinding(b) == RiskLevelHigh))
		if !needsGate {
			continue
		}
		gateID := trustGateID(b.Action)
		if _, ok := has[gateID]; ok {
			continue
		}
		add(GraphNode{
			ID:    gateID,
			Type:  NodeTypeTrustVerification,
			Title: fmt.Sprintf("Trust gate for %s", strings.ReplaceAll(b.Action, "_", " ")),
			Agent: "local",
			Risk:  RiskLevelHigh,
			RequiredChecks: []string{
				"trust",
				"flows",
			},
		})
		edges = append(edges, GraphEdge{
			From:   b.NodeID,
			To:     gateID,
			Type:   EdgeTypeMustRunAfter,
			Reason: "trust gate required for public contract or high-risk change",
		})
	}

	if !hasObservability {
		for i := range nodes {
			if nodes[i].Type == NodeTypeValidation && strings.Contains(nodes[i].ID, "verify-") {
				nodes[i].Risk = maxRisk(nodes[i].Risk, RiskLevelMedium)
			}
		}
	}

	return nodes, edges
}

func isPublicContract(ref string) bool {
	ref = strings.TrimSpace(ref)
	return ref != "" && !strings.HasPrefix(ref, "TODO:")
}

func trustGateID(action string) string {
	action = strings.TrimSpace(strings.ToLower(action))
	action = strings.ReplaceAll(action, "_", "-")
	if action == "" {
		return "trust-gate"
	}
	return "trust-gate-" + action
}

func bindingForNode(bindings []TaskBinding, nodeID string) *TaskBinding {
	for i := range bindings {
		if bindings[i].NodeID == nodeID {
			return &bindings[i]
		}
	}
	return nil
}
