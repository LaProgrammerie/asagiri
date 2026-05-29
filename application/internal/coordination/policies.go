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
					result.Warnings = append(result.Warnings, fmt.Sprintf(
						"flow %q may require security review (configured: %q)", graph.Flow, sensitive,
					))
				}
			}
		}
	}
	return result
}

func hasReviewNode(graph ExecutionGraph) bool {
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeReview {
			return true
		}
	}
	return false
}
