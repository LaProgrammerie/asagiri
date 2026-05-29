package executiongraph

import (
	"sort"
	"strings"
)

// RiskAssessment captures planning risk metadata for a node (spec §14).
type RiskAssessment struct {
	Level                 RiskLevel
	BlastRadius           int
	Reasons               []string
	RequiredChecks        []string
	RequiresHumanApproval bool
}

var planningEdgeTypes = map[EdgeType]bool{
	EdgeTypeRequires:              true,
	EdgeTypeMustRunAfter:          true,
	EdgeTypeBlocks:                true,
	EdgeTypeRollbackDependsOn:     true,
	EdgeTypeProducesContextFor:    true,
	EdgeTypeValidates:             true,
	EdgeTypeRequiresHumanApproval: true,
}

// AssessNodeRisk computes risk metadata for a single node.
func AssessNodeRisk(node GraphNode, binding *TaskBinding, edges []GraphEdge) RiskAssessment {
	level := node.Risk
	if level == "" {
		level = defaultRiskForType(node.Type)
	}
	reasons := riskReasons(node, binding)
	level = elevateRisk(level, reasons, binding)

	checks := requiredChecksFor(node, binding)
	approval := requiresHumanApproval(level, node.Type, binding)

	return RiskAssessment{
		Level:                 level,
		BlastRadius:           computeBlastRadius(node.ID, edges),
		Reasons:               reasons,
		RequiredChecks:        checks,
		RequiresHumanApproval: approval,
	}
}

// ApplyRiskEnrichment updates node risk fields and returns optional approval nodes/edges.
func ApplyRiskEnrichment(nodes []GraphNode, bindings []TaskBinding, edges []GraphEdge) ([]GraphNode, []GraphEdge) {
	bindingByNode := bindingIndex(bindings)
	out := make([]GraphNode, 0, len(nodes)+1)
	for _, n := range nodes {
		var binding *TaskBinding
		if b, ok := bindingByNode[n.ID]; ok {
			binding = &b
		}
		assessment := AssessNodeRisk(n, binding, edges)
		n.Risk = assessment.Level
		n.BlastRadius = assessment.BlastRadius
		n.RequiredChecks = assessment.RequiredChecks
		n.RequiresHumanApproval = assessment.RequiresHumanApproval
		out = append(out, n)
	}

	additionalEdges := make([]GraphEdge, 0)
	if approvalNode, approvalEdges := ensureManualApprovalNode(out, edges); approvalNode != nil {
		out = append(out, *approvalNode)
		additionalEdges = append(additionalEdges, approvalEdges...)
	}
	return out, additionalEdges
}

// HighestRisk returns the maximum risk level across nodes.
func HighestRisk(nodes []GraphNode) RiskLevel {
	highest := RiskLevelLow
	for _, n := range nodes {
		if riskRank(n.Risk) > riskRank(highest) {
			highest = n.Risk
		}
	}
	return highest
}

func defaultRiskForType(nodeType NodeType) RiskLevel {
	switch nodeType {
	case NodeTypeInvestigation:
		return RiskLevelLow
	case NodeTypeManualApproval:
		return RiskLevelCritical
	case NodeTypeReview, NodeTypeTrustVerification, NodeTypeRollback:
		return RiskLevelHigh
	case NodeTypeValidation, NodeTypeContractGeneration, NodeTypeArchitectureDerivation:
		return RiskLevelMedium
	default:
		return RiskLevelMedium
	}
}

func riskReasons(node GraphNode, binding *TaskBinding) []string {
	reasons := make([]string, 0, 4)
	if binding != nil && binding.Sensitive {
		reasons = append(reasons, "touches sensitive action")
	}
	if binding != nil && strings.Contains(strings.ToLower(binding.Action), "permission") {
		reasons = append(reasons, "touches permission model")
	}
	if binding != nil && binding.ContractRef != "" && !strings.HasPrefix(binding.ContractRef, "TODO:") {
		reasons = append(reasons, "emits public contract change")
	}
	if strings.Contains(strings.ToLower(node.Title), "migration") ||
		strings.Contains(strings.ToLower(node.ID), "migration") {
		reasons = append(reasons, "database migration")
	}
	if node.Type == NodeTypeTrustVerification {
		reasons = append(reasons, "requires trust gate")
	}
	sort.Strings(reasons)
	return dedupeStrings(reasons)
}

func elevateRisk(base RiskLevel, reasons []string, binding *TaskBinding) RiskLevel {
	level := base
	for _, reason := range reasons {
		switch {
		case strings.Contains(reason, "permission"):
			level = maxRisk(level, RiskLevelCritical)
		case strings.Contains(reason, "sensitive"):
			level = maxRisk(level, RiskLevelHigh)
		case strings.Contains(reason, "public contract"):
			level = maxRisk(level, RiskLevelHigh)
		case strings.Contains(reason, "migration"):
			level = maxRisk(level, RiskLevelHigh)
		}
	}
	if binding != nil && binding.Sensitive && riskRank(level) < riskRank(RiskLevelHigh) {
		level = RiskLevelHigh
	}
	return level
}

func requiredChecksFor(node GraphNode, binding *TaskBinding) []string {
	checks := append([]string(nil), node.RequiredChecks...)
	switch node.Type {
	case NodeTypeImplementation:
		checks = appendUnique(checks, "tests")
		if binding != nil && binding.Sensitive {
			checks = appendUnique(checks, "security", "observability")
		}
	case NodeTypeReview:
		checks = appendUnique(checks, "security")
	case NodeTypeTrustVerification:
		checks = appendUnique(checks, "trust", "flows")
	case NodeTypeValidation:
		if binding != nil && binding.ContractRef != "" && !strings.HasPrefix(binding.ContractRef, "TODO:") {
			checks = appendUnique(checks, "backward_compatibility")
		} else if strings.Contains(node.ID, "contracts") {
			checks = appendUnique(checks, "backward_compatibility")
		} else {
			checks = appendUnique(checks, "flows")
		}
	case NodeTypeContractGeneration:
		checks = appendUnique(checks, "contracts")
	}
	sort.Strings(checks)
	return checks
}

func requiresHumanApproval(level RiskLevel, nodeType NodeType, binding *TaskBinding) bool {
	if nodeType == NodeTypeManualApproval {
		return true
	}
	if level == RiskLevelCritical {
		return true
	}
	if binding != nil && binding.Sensitive && strings.Contains(strings.ToLower(binding.Action), "permission") {
		return true
	}
	if binding != nil && strings.Contains(strings.ToLower(binding.Action), "migration") {
		return true
	}
	return false
}

func computeBlastRadius(nodeID string, edges []GraphEdge) int {
	visited := map[string]struct{}{nodeID: {}}
	queue := []string{nodeID}
	count := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, e := range edges {
			if e.From != current || !planningEdgeTypes[e.Type] {
				continue
			}
			if _, seen := visited[e.To]; seen {
				continue
			}
			visited[e.To] = struct{}{}
			count++
			queue = append(queue, e.To)
		}
	}
	return count
}

func ensureManualApprovalNode(nodes []GraphNode, edges []GraphEdge) (*GraphNode, []GraphEdge) {
	hasApprovalNode := false
	needsApproval := false
	var sourceID string
	for _, n := range nodes {
		if n.Type == NodeTypeManualApproval {
			hasApprovalNode = true
		}
		if n.RequiresHumanApproval && n.Type != NodeTypeManualApproval {
			needsApproval = true
			if n.Type == NodeTypeTrustVerification || n.Type == NodeTypeReview {
				sourceID = n.ID
			}
		}
	}
	if !needsApproval || hasApprovalNode {
		return nil, nil
	}
	if sourceID == "" {
		for _, n := range nodes {
			if n.RequiresHumanApproval && n.Type != NodeTypeManualApproval {
				sourceID = n.ID
				break
			}
		}
	}
	node := GraphNode{
		ID:                    "manual-approval",
		Type:                  NodeTypeManualApproval,
		Title:                 "Human approval for high-risk change",
		Agent:                 "local",
		Risk:                  RiskLevelCritical,
		RequiresHumanApproval: true,
	}
	edge := GraphEdge{
		From:   sourceID,
		To:     node.ID,
		Type:   EdgeTypeRequiresHumanApproval,
		Reason: "high-risk change requires human approval",
	}
	return &node, []GraphEdge{edge}
}

func riskRank(r RiskLevel) int {
	switch r {
	case RiskLevelLow:
		return 1
	case RiskLevelMedium:
		return 2
	case RiskLevelHigh:
		return 3
	case RiskLevelCritical:
		return 4
	default:
		return 0
	}
}

func maxRisk(a, b RiskLevel) RiskLevel {
	if riskRank(a) >= riskRank(b) {
		return a
	}
	return b
}

func appendUnique(values []string, items ...string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		seen[v] = struct{}{}
	}
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		values = append(values, item)
	}
	return values
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	out := values[:1]
	for i := 1; i < len(values); i++ {
		if values[i] == out[len(out)-1] {
			continue
		}
		out = append(out, values[i])
	}
	return out
}
