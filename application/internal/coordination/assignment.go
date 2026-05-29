package coordination

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

// ExecutionGraph aliases the execution graph model (spec-my-D §18).
type ExecutionGraph = executiongraph.ExecutionGraph

// GraphNode aliases a graph node for assignment APIs.
type GraphNode = executiongraph.GraphNode

// AgentRole names a specialized agent role (spec-my-D §3–4).
type AgentRole string

const (
	RoleInvestigator          AgentRole = "investigator"
	RoleArchitect             AgentRole = "architect"
	RolePlanner               AgentRole = "planner"
	RoleImplementer           AgentRole = "implementer"
	RoleReviewer              AgentRole = "reviewer"
	RoleValidator             AgentRole = "validator"
	RoleSecurityAuditor       AgentRole = "security_auditor"
	RolePerformanceAuditor    AgentRole = "performance_auditor"
	RoleObservabilityAuditor  AgentRole = "observability_auditor"
	RoleDocumenter            AgentRole = "documenter"
)

// IsolationMode controls how an agent executes relative to the workspace (spec-my-D §5).
type IsolationMode string

const (
	IsolationShared           IsolationMode = "shared"
	IsolationIsolatedWorktree IsolationMode = "isolated_worktree"
	IsolationReadonly         IsolationMode = "readonly"
	IsolationSandbox          IsolationMode = "sandbox"
)

// AgentProfile binds a coordination profile to a configured agents entry (spec-my-D §4).
type AgentProfile struct {
	ID               string        `yaml:"id,omitempty" json:"id,omitempty"`
	Role             AgentRole     `yaml:"role" json:"role"`
	Agent            string        `yaml:"agent" json:"agent"`
	Capabilities     []string      `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Restrictions     []string      `yaml:"restrictions,omitempty" json:"restrictions,omitempty"`
	MaxContextTokens int           `yaml:"max_context_tokens,omitempty" json:"max_context_tokens,omitempty"`
	Isolation        IsolationMode `yaml:"isolation,omitempty" json:"isolation,omitempty"`
}

// AgentAssignment is the outcome of assigning an agent to a graph node.
type AgentAssignment struct {
	NodeID    string        `json:"node_id"`
	AgentRef  string        `json:"agent_ref"`
	Role      AgentRole     `json:"role"`
	Isolation IsolationMode `json:"isolation"`
	ProfileID string        `json:"profile_id,omitempty"`
}

// AgentAssigner selects agent, role, and isolation for a node.
type AgentAssigner interface {
	Assign(ctx context.Context, node GraphNode) (AgentAssignment, error)
}

// RoleForNodeType maps execution graph node types to coordination roles.
func RoleForNodeType(nodeType executiongraph.NodeType) AgentRole {
	switch nodeType {
	case executiongraph.NodeTypeInvestigation:
		return RoleInvestigator
	case executiongraph.NodeTypeArchitectureDerivation, executiongraph.NodeTypeContractGeneration:
		return RoleArchitect
	case executiongraph.NodeTypeEnrichment:
		return RolePlanner
	case executiongraph.NodeTypeImplementation:
		return RoleImplementer
	case executiongraph.NodeTypeReview:
		return RoleReviewer
	case executiongraph.NodeTypeValidation, executiongraph.NodeTypeTrustVerification,
		executiongraph.NodeTypeReleaseCheck:
		return RoleValidator
	case executiongraph.NodeTypeDocumentation:
		return RoleDocumenter
	case executiongraph.NodeTypeManualApproval:
		return RolePlanner
	case executiongraph.NodeTypeRollback:
		return RoleInvestigator
	default:
		return RoleImplementer
	}
}

// ValidateRole returns an error when role is not a known AgentRole.
func ValidateRole(role AgentRole) error {
	switch role {
	case RoleInvestigator, RoleArchitect, RolePlanner, RoleImplementer, RoleReviewer,
		RoleValidator, RoleSecurityAuditor, RolePerformanceAuditor, RoleObservabilityAuditor,
		RoleDocumenter:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidRole, role)
	}
}

// ValidateIsolation returns an error when mode is not a known IsolationMode.
func ValidateIsolation(mode IsolationMode) error {
	switch mode {
	case IsolationShared, IsolationIsolatedWorktree, IsolationReadonly, IsolationSandbox:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidIsolation, mode)
	}
}

// AssignerConfig feeds DefaultAssigner (coordination block + agent registry keys).
type AssignerConfig struct {
	DefaultIsolation IsolationMode
	Assignment       map[string]string
	Profiles         map[string]AgentProfile
}

// DefaultAssigner delegates agent ref selection to executiongraph.DefaultAgentFor.
type DefaultAssigner struct {
	Config AssignerConfig
}

// Assign resolves agent ref, role, and isolation for a node.
func (a *DefaultAssigner) Assign(_ context.Context, node GraphNode) (AgentAssignment, error) {
	if node.ID == "" {
		return AgentAssignment{}, fmt.Errorf("%w: node id required", ErrInvalidAssignment)
	}
	cfg := a.Config
	role := RoleForNodeType(node.Type)
	if err := ValidateRole(role); err != nil {
		return AgentAssignment{}, err
	}
	isolation := cfg.DefaultIsolation
	if err := ValidateIsolation(isolation); err != nil {
		return AgentAssignment{}, err
	}

	agentRef := executiongraph.DefaultAgentFor(node, nil)
	if key := assignmentKey(node); key != "" {
		if override, ok := cfg.Assignment[key]; ok && strings.TrimSpace(override) != "" {
			agentRef = strings.TrimSpace(override)
		}
	}

	profileID := ""
	for id, p := range cfg.Profiles {
		if p.Role == role && p.Agent == agentRef {
			profileID = id
			if p.Isolation != "" {
				if err := ValidateIsolation(p.Isolation); err != nil {
					return AgentAssignment{}, err
				}
				isolation = p.Isolation
			}
			break
		}
	}

	return AgentAssignment{
		NodeID:    node.ID,
		AgentRef:  agentRef,
		Role:      role,
		Isolation: isolation,
		ProfileID: profileID,
	}, nil
}

func assignmentKey(node GraphNode) string {
	switch node.Type {
	case executiongraph.NodeTypeInvestigation:
		return "investigation"
	case executiongraph.NodeTypeArchitectureDerivation:
		return "architecture_review"
	case executiongraph.NodeTypeImplementation:
		switch node.Risk {
		case executiongraph.RiskLevelHigh, executiongraph.RiskLevelCritical:
			return "implementation.high"
		default:
			return "implementation.medium"
		}
	case executiongraph.NodeTypeReview:
		return "security_review"
	case executiongraph.NodeTypeValidation, executiongraph.NodeTypeTrustVerification:
		return "validation"
	default:
		return ""
	}
}
