package executiongraph

// AgentAssigner selects an execution agent for a graph node (spec §12).
type AgentAssigner interface {
	Assign(node GraphNode, binding *TaskBinding) string
}

// DefaultAgentAssigner maps node type and risk to a default agent profile.
type DefaultAgentAssigner struct{}

func (DefaultAgentAssigner) Assign(node GraphNode, binding *TaskBinding) string {
	return DefaultAgentFor(node, binding)
}

// DefaultAgentFor returns the default agent for a node based on type and risk.
func DefaultAgentFor(node GraphNode, binding *TaskBinding) string {
	switch node.Type {
	case NodeTypeInvestigation, NodeTypeValidation, NodeTypeTrustVerification,
		NodeTypeManualApproval, NodeTypeRollback, NodeTypeReleaseCheck,
		NodeTypeArchitectureDerivation, NodeTypeContractGeneration:
		return "local"
	case NodeTypeEnrichment, NodeTypeDocumentation:
		return "ollama"
	case NodeTypeReview:
		if riskRank(node.Risk) >= riskRank(RiskLevelHigh) {
			return "codex"
		}
		return "local"
	case NodeTypeImplementation:
		if binding != nil && binding.Sensitive {
			return "claude"
		}
		switch node.Risk {
		case RiskLevelCritical, RiskLevelHigh:
			return "claude"
		default:
			return "cursor"
		}
	default:
		return "local"
	}
}

// AssignAgents fills empty or default agent fields on nodes.
func AssignAgents(nodes []GraphNode, bindings []TaskBinding) []GraphNode {
	bindingByNode := bindingIndex(bindings)
	out := make([]GraphNode, len(nodes))
	for i, n := range nodes {
		var binding *TaskBinding
		if b, ok := bindingByNode[n.ID]; ok {
			binding = &b
		}
		n.Agent = DefaultAgentFor(n, binding)
		out[i] = n
	}
	return out
}

func bindingIndex(bindings []TaskBinding) map[string]TaskBinding {
	out := make(map[string]TaskBinding, len(bindings))
	for _, b := range bindings {
		if b.NodeID != "" {
			out[b.NodeID] = b
		}
	}
	return out
}
