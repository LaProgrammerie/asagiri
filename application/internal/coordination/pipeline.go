package coordination

import (
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

// DefaultPipelineRoles is the canonical multi-agent order (spec-my-D §7).
var DefaultPipelineRoles = []AgentRole{
	RoleInvestigator,
	RoleArchitect,
	RoleImplementer,
	RoleReviewer,
	RoleValidator,
}

// PipelineStep groups graph nodes mapped to one pipeline role.
type PipelineStep struct {
	Role    AgentRole            `json:"role"`
	Order   int                  `json:"order"`
	NodeIDs []string             `json:"node_ids,omitempty"`
	Results []PipelineStepResult `json:"results,omitempty"`
}

// PipelineStepResult holds structured output for one node in a pipeline step (spec-my-D §7).
type PipelineStepResult struct {
	NodeID     string    `json:"node_id"`
	Role       AgentRole `json:"role"`
	Outputs    []string  `json:"outputs,omitempty"`
	Confidence float64   `json:"confidence,omitempty"`
	CostEUR    float64   `json:"cost_eur,omitempty"`
	Tokens     int       `json:"tokens,omitempty"`
}

// DefaultPipeline maps execution graph nodes onto a replayable role-ordered pipeline.
type DefaultPipeline struct {
	Roles []AgentRole
}

// NewDefaultPipeline builds a pipeline from config roles or defaults.
func NewDefaultPipeline(cfg *config.Config) *DefaultPipeline {
	roles := DefaultPipelineRoles
	if cfg != nil && len(cfg.Coordination.Pipeline) > 0 {
		roles = make([]AgentRole, 0, len(cfg.Coordination.Pipeline))
		for _, r := range cfg.Coordination.Pipeline {
			role := AgentRole(strings.TrimSpace(r))
			if err := ValidateRole(role); err == nil {
				roles = append(roles, role)
			}
		}
		if len(roles) == 0 {
			roles = DefaultPipelineRoles
		}
	}
	return &DefaultPipeline{Roles: roles}
}

// MapGraph assigns each node to a pipeline step in deterministic replay order.
func (p *DefaultPipeline) MapGraph(graph ExecutionGraph) []PipelineStep {
	roles := DefaultPipelineRoles
	if p != nil && len(p.Roles) > 0 {
		roles = p.Roles
	}

	roleIndex := make(map[AgentRole]int, len(roles))
	stepsByRole := make(map[AgentRole]*PipelineStep, len(roles))
	for i, role := range roles {
		roleIndex[role] = i
		stepsByRole[role] = &PipelineStep{Role: role, Order: i}
	}

	nodes := append([]executiongraph.GraphNode(nil), graph.Nodes...)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	for _, node := range nodes {
		role := RoleForNodeType(node.Type)
		idx, ok := roleIndex[role]
		if !ok {
			continue
		}
		step := stepsByRole[role]
		step.NodeIDs = append(step.NodeIDs, node.ID)
		step.Results = append(step.Results, stepResultFromNode(node, role))
		_ = idx
	}

	out := make([]PipelineStep, 0, len(roles))
	for _, role := range roles {
		if step, ok := stepsByRole[role]; ok && len(step.NodeIDs) > 0 {
			out = append(out, *step)
		}
	}
	return out
}

func stepResultFromNode(node executiongraph.GraphNode, role AgentRole) PipelineStepResult {
	confidence := 0.85
	if node.Status == executiongraph.NodeStatusFailed {
		confidence = 0.2
	}
	return PipelineStepResult{
		NodeID:     node.ID,
		Role:       role,
		Outputs:    append([]string(nil), node.Outputs...),
		Confidence: confidence,
		CostEUR:    node.EstimatedCost,
		Tokens:     estimateTokensFromOutputs(node.Outputs),
	}
}

func estimateTokensFromOutputs(paths []string) int {
	const charsPerToken = 4
	n := 0
	for _, p := range paths {
		n += len(p) / charsPerToken
	}
	if n == 0 && len(paths) > 0 {
		return len(paths)
	}
	return n
}
