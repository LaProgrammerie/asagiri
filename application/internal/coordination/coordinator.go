package coordination

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// AgentCoordinator orchestrates agent assignment and policies for an execution graph.
type AgentCoordinator interface {
	Coordinate(ctx context.Context, graph ExecutionGraph) (CoordinationResult, error)
}

// CoordinationResult holds the last coordinated graph and assignments.
type CoordinationResult struct {
	Graph       ExecutionGraph
	Assignments []AgentAssignment
	Policy      PolicyResult
	Pipeline    []PipelineStep
}

// DefaultCoordinator validates the graph, evaluates policies, and assigns agents to nodes.
type DefaultCoordinator struct {
	Assigner  AgentAssigner
	Policies  PolicyEvaluator
	Emitter   *CoordinationEmitter
	Handoff   HandoffBuilder
	Pipeline  *DefaultPipeline
	RepoRoot  string
	last      CoordinationResult
}

// LastResult returns the outcome of the most recent successful Coordinate call.
// Deprecated: prefer the CoordinationResult return value from Coordinate.
func (c *DefaultCoordinator) LastResult() CoordinationResult {
	if c == nil {
		return CoordinationResult{}
	}
	return c.last
}

// Coordinate validates the graph, applies policies, and assigns each node.
func (c *DefaultCoordinator) Coordinate(ctx context.Context, graph ExecutionGraph) (CoordinationResult, error) {
	if c == nil {
		return CoordinationResult{}, fmt.Errorf("%w: coordinator is nil", ErrInvalidGraph)
	}
	if err := graph.Validate(); err != nil {
		return CoordinationResult{}, fmt.Errorf("%w: %v", ErrInvalidGraph, err)
	}

	policy := c.Policies.Evaluate(graph)
	if !policy.OK {
		return CoordinationResult{}, fmt.Errorf("%w: %s", ErrPolicyViolation, policy.Errors[0])
	}

	assigner := c.Assigner
	if assigner == nil {
		assigner = &DefaultAssigner{}
	}

	assignments := make([]AgentAssignment, 0, len(graph.Nodes))
	nodes := make([]executiongraph.GraphNode, len(graph.Nodes))
	var knowledgeStore knowledge.GraphStore
	if strings.TrimSpace(c.RepoRoot) != "" {
		store, err := OpenKnowledgeStoreIfPresent(c.RepoRoot)
		if err != nil {
			return CoordinationResult{}, err
		}
		if store != nil {
			knowledgeStore = store
			defer store.Close()
		}
	}

	for i, node := range graph.Nodes {
		asg, err := assigner.Assign(ctx, node)
		if err != nil {
			return CoordinationResult{}, err
		}
		if knowledgeStore != nil && graph.Flow != "" {
			asg, err = ApplyGraphAgentRouting(ctx, knowledgeStore, graph.Flow, "", node, asg)
			if err != nil {
				return CoordinationResult{}, err
			}
		}
		assignments = append(assignments, asg)
		node.Agent = asg.AgentRef
		nodes[i] = node
		if c.Emitter != nil {
			_ = c.Emitter.EmitAgentStarted(graph.ID, graph.Flow, asg)
		}
	}
	graph.Nodes = nodes

	postPolicy := c.Policies.EvaluateWithAssignments(graph, assignments)
	if !postPolicy.OK {
		return CoordinationResult{}, fmt.Errorf("%w: %s", ErrPolicyViolation, postPolicy.Errors[0])
	}

	var pipelineSteps []PipelineStep
		if c.Pipeline != nil {
		pipelineSteps = c.Pipeline.MapGraph(graph)
		if c.Handoff != nil {
			_ = c.buildPipelineHandoffs(ctx, graph, pipelineSteps, assignments, knowledgeStore)
		}
	}

	if strings.TrimSpace(c.RepoRoot) != "" {
		cg := BuildCoordinationGraph(graph, assignments)
		_, _ = PersistCoordinationGraph(c.RepoRoot, cg)
	}

	result := CoordinationResult{
		Graph:       graph,
		Assignments: assignments,
		Policy:      postPolicy,
		Pipeline:    pipelineSteps,
	}
	c.last = result
	return result, nil
}

func (c *DefaultCoordinator) buildPipelineHandoffs(
	ctx context.Context,
	graph ExecutionGraph,
	steps []PipelineStep,
	assignments []AgentAssignment,
	knowledgeStore knowledge.GraphStore,
) error {
	if c.Handoff == nil {
		return nil
	}
	nodeByID := make(map[string]executiongraph.GraphNode, len(graph.Nodes))
	for _, n := range graph.Nodes {
		nodeByID[n.ID] = n
	}

	asgByNode := make(map[string]AgentAssignment, len(assignments))
	for _, a := range assignments {
		asgByNode[a.NodeID] = a
	}
	for i := 1; i < len(steps); i++ {
		prev, cur := steps[i-1], steps[i]
		if len(prev.Results) == 0 || len(cur.NodeIDs) == 0 {
			continue
		}
		last := prev.Results[len(prev.Results)-1]
		targetNode := cur.NodeIDs[0]
		files := append([]string(nil), last.Outputs...)
		if node, ok := nodeByID[last.NodeID]; ok {
			pack, _, _ := ReduceContext(node, graph, HandoffHints{
				KnowledgeStore: knowledgeStore,
				GraphFlow:      graph.Flow,
				MaxFiles:       64,
			})
			files = uniqueFilePaths(append(files, pack.Files...))
		}
		h, err := c.Handoff.Build(ctx, AgentResult{
			NodeID:     last.NodeID,
			Role:       prev.Role,
			AgentRef:   asgByNode[last.NodeID].AgentRef,
			Summary:    fmt.Sprintf("pipeline handoff %s → %s", prev.Role, cur.Role),
			Files:      files,
			TargetRole: cur.Role,
			Confidence: last.Confidence,
		})
		if err != nil {
			return err
		}
		if knowledgeStore != nil && graph.Flow != "" {
			h, err = EnrichHandoffWithGraph(ctx, knowledgeStore, graph.Flow, "", h)
			if err != nil {
				return err
			}
		}
		if c.Emitter != nil {
			_ = c.Emitter.EmitHandoffCreated(graph.ID, graph.Flow, h)
		}
		_ = targetNode
	}
	return nil
}

func uniqueFilePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
