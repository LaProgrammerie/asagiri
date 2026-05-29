package coordination

import (
	"context"
	"fmt"
)

// AgentCoordinator orchestrates agent assignment and policies for an execution graph.
type AgentCoordinator interface {
	Coordinate(ctx context.Context, graph ExecutionGraph) error
}

// CoordinationResult holds the last coordinated graph and assignments.
type CoordinationResult struct {
	Graph       ExecutionGraph
	Assignments []AgentAssignment
	Policy      PolicyResult
}

// DefaultCoordinator validates the graph, evaluates policies, and assigns agents to nodes.
type DefaultCoordinator struct {
	Assigner  AgentAssigner
	Policies  PolicyEvaluator
	Emitter   *CoordinationEmitter
	RepoRoot  string
	last      CoordinationResult
}

// LastResult returns the outcome of the most recent Coordinate call.
func (c *DefaultCoordinator) LastResult() CoordinationResult {
	if c == nil {
		return CoordinationResult{}
	}
	return c.last
}

// Coordinate validates the graph, applies policies, and assigns each node.
func (c *DefaultCoordinator) Coordinate(ctx context.Context, graph ExecutionGraph) error {
	if c == nil {
		return fmt.Errorf("%w: coordinator is nil", ErrInvalidGraph)
	}
	if err := graph.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidGraph, err)
	}

	policy := c.Policies.Evaluate(graph)
	if !policy.OK {
		return fmt.Errorf("%w: %s", ErrPolicyViolation, policy.Errors[0])
	}

	assigner := c.Assigner
	if assigner == nil {
		assigner = &DefaultAssigner{}
	}

	assignments := make([]AgentAssignment, 0, len(graph.Nodes))
	nodes := make([]GraphNode, len(graph.Nodes))
	for i, node := range graph.Nodes {
		asg, err := assigner.Assign(ctx, node)
		if err != nil {
			return err
		}
		assignments = append(assignments, asg)
		node.Agent = asg.AgentRef
		nodes[i] = node
		if c.Emitter != nil {
			_ = c.Emitter.EmitAgentStarted(graph.ID, graph.Flow, asg)
		}
	}
	graph.Nodes = nodes

	c.last = CoordinationResult{
		Graph:       graph,
		Assignments: assignments,
		Policy:      policy,
	}
	return nil
}
