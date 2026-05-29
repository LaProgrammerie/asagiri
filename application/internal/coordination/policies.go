package coordination

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

// CoordinationPolicies mirrors coordination: policy fields (spec-my-D §11).
type CoordinationPolicies struct {
	MaxParallelAgents          int
	RequireIndependentReview   bool
	AllowSelfReview            bool
	RequireSecurityReviewFor   []string
	DefaultIsolation           IsolationMode
}

// PolicyEvaluator checks coordination policies against a graph.
type PolicyEvaluator struct {
	Policies CoordinationPolicies
}

// PolicyResult summarizes policy evaluation.
type PolicyResult struct {
	OK       bool
	Warnings []string
	Errors   []string
}

// Evaluate runs structural and policy checks for coordination.
func (e *PolicyEvaluator) Evaluate(graph ExecutionGraph) PolicyResult {
	return e.evaluate(graph, nil)
}

// EvaluateWithAssignments runs structural checks plus cross-agent validation (spec-my-D §9, §11).
func (e *PolicyEvaluator) EvaluateWithAssignments(graph ExecutionGraph, assignments []AgentAssignment) PolicyResult {
	return e.evaluate(graph, assignments)
}

func (e *PolicyEvaluator) evaluate(graph ExecutionGraph, assignments []AgentAssignment) PolicyResult {
	result := PolicyResult{OK: true}
	if err := graph.Validate(); err != nil {
		result.OK = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	if e == nil {
		return result
	}
	p := e.Policies
	if p.MaxParallelAgents > 0 && graph.Strategy.MaxParallel > p.MaxParallelAgents {
		result.OK = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"graph max_parallel %d exceeds coordination max_parallel_agents %d",
			graph.Strategy.MaxParallel, p.MaxParallelAgents,
		))
	}
	if graph.Flow != "" && len(p.RequireSecurityReviewFor) > 0 {
		for _, sensitive := range p.RequireSecurityReviewFor {
			if strings.Contains(strings.ToLower(graph.Flow), strings.ToLower(sensitive)) {
				if !hasReviewNode(graph) {
					msg := fmt.Sprintf(
						"flow %q requires security review (configured: %q)", graph.Flow, sensitive,
					)
					result.OK = false
					result.Errors = append(result.Errors, msg)
				}
			}
		}
	}
	if len(assignments) > 0 {
		cross := e.evaluateCrossAgent(graph, assignments)
		if !cross.OK {
			result.OK = false
			result.Errors = append(result.Errors, cross.Errors...)
		}
		result.Warnings = append(result.Warnings, cross.Warnings...)
	}
	return result
}

func (e *PolicyEvaluator) evaluateCrossAgent(graph ExecutionGraph, assignments []AgentAssignment) PolicyResult {
	result := PolicyResult{OK: true}
	if e == nil {
		return result
	}
	p := e.Policies
	if !p.RequireIndependentReview || p.AllowSelfReview {
		return result
	}

	implAgents := agentRefsForRoles(graph, assignments, RoleImplementer)
	reviewerAgents := agentRefsForRoles(graph, assignments, RoleReviewer)
	validatorAgents := agentRefsForRoles(graph, assignments, RoleValidator)

	for impl := range implAgents {
		if _, ok := reviewerAgents[impl]; ok {
			result.OK = false
			result.Errors = append(result.Errors, fmt.Sprintf(
				"reviewer must differ from implementer: agent_ref %q", impl,
			))
		}
		if _, ok := validatorAgents[impl]; ok {
			result.OK = false
			result.Errors = append(result.Errors, fmt.Sprintf(
				"validator must differ from implementer: agent_ref %q", impl,
			))
		}
	}
	return result
}

func agentRefsForRoles(graph ExecutionGraph, assignments []AgentAssignment, role AgentRole) map[string]struct{} {
	byNode := make(map[string]AgentAssignment, len(assignments))
	for _, a := range assignments {
		byNode[a.NodeID] = a
	}
	out := make(map[string]struct{})
	for _, n := range graph.Nodes {
		if RoleForNodeType(n.Type) != role {
			continue
		}
		ref := n.Agent
		if a, ok := byNode[n.ID]; ok && a.AgentRef != "" {
			ref = a.AgentRef
		}
		if ref != "" {
			out[ref] = struct{}{}
		}
	}
	return out
}

func hasReviewNode(graph ExecutionGraph) bool {
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeReview {
			return true
		}
	}
	return false
}
