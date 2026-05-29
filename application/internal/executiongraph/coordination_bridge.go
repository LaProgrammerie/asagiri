package executiongraph

import "context"

// CoordinationAssignment is the runner-facing assignment record (avoids import cycle with coordination).
type CoordinationAssignment struct {
	NodeID    string `json:"node_id"`
	AgentRef  string `json:"agent_ref"`
	Role      string `json:"role"`
	Isolation string `json:"isolation,omitempty"`
	ProfileID string `json:"profile_id,omitempty"`
}

// CoordinationResult is returned by GraphCoordinator before node execution.
type CoordinationResult struct {
	Graph       ExecutionGraph
	Assignments []CoordinationAssignment
}

// GraphCoordinator coordinates agent assignment before graph execution (spec-my-D §18).
type GraphCoordinator func(ctx context.Context, graph ExecutionGraph) (CoordinationResult, error)

// RoleForNodeType maps graph node types to coordination role names (spec-my-D §3–4).
func RoleForNodeType(nodeType NodeType) string {
	switch nodeType {
	case NodeTypeInvestigation:
		return "investigator"
	case NodeTypeArchitectureDerivation, NodeTypeContractGeneration:
		return "architect"
	case NodeTypeEnrichment:
		return "planner"
	case NodeTypeImplementation:
		return "implementer"
	case NodeTypeReview:
		return "reviewer"
	case NodeTypeValidation, NodeTypeTrustVerification, NodeTypeReleaseCheck:
		return "validator"
	case NodeTypeDocumentation:
		return "documenter"
	case NodeTypeManualApproval:
		return "planner"
	case NodeTypeRollback:
		return "investigator"
	default:
		return "implementer"
	}
}
