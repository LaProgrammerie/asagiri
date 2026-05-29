package executiongraph

import (
	"context"
	"fmt"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

// RunOptions configures graph execution (spec §5.3, §19).
type RunOptions struct {
	DryRun      bool
	CIMode      bool
	StrictTrust bool
	Scheduler   GraphScheduler
	Gates       trust.GateEvaluator
}

// DefaultRunner performs graph execution with runtime events and checkpoints (spec §19–20).
type DefaultRunner struct {
	Repo     *Repository
	RepoRoot string
	Events   *runtime.GraphEmitter
}

// NewRunner returns a runner backed by repoRoot storage.
func NewRunner(repoRoot string) *DefaultRunner {
	store, err := runtime.Open(repoRoot)
	var emitter *runtime.GraphEmitter
	if err == nil {
		emitter = &runtime.GraphEmitter{Store: store}
	}
	return &DefaultRunner{
		Repo:     NewRepository(repoRoot),
		RepoRoot: repoRoot,
		Events:   emitter,
	}
}

// DryRun plans, schedules, persists artefacts, and marks the graph ready without agent execution.
func (r *DefaultRunner) DryRun(ctx context.Context, graph ExecutionGraph, ciMode bool) (GraphRunResult, ExecutionSchedule, GraphArtifacts, error) {
	sched := r.scheduler()
	schedule, err := sched.Schedule(ctx, ScheduleRequest{Graph: graph, CIMode: ciMode})
	if err != nil {
		return GraphRunResult{}, ExecutionSchedule{}, GraphArtifacts{}, err
	}

	if err := TransitionGraph(graph.Status, GraphStatusReady, false); err == nil {
		graph.Status = GraphStatusReady
	}

	artifacts, err := r.Repo.SaveAll(graph, &schedule)
	if err != nil {
		return GraphRunResult{}, ExecutionSchedule{}, GraphArtifacts{}, err
	}

	r.emit(runtime.EventGraphCreated, graph, "", map[string]any{"status": graph.Status, "dry_run": true})

	return GraphRunResult{
		GraphID: graph.ID,
		Status:  graph.Status,
	}, schedule, artifacts, nil
}

// Run executes or simulates graph nodes according to options (spec §19).
func (r *DefaultRunner) Run(ctx context.Context, graph ExecutionGraph, schedule ExecutionSchedule, opts RunOptions) (GraphRunResult, error) {
	if opts.DryRun {
		result, _, _, err := r.DryRun(ctx, graph, opts.CIMode)
		return result, err
	}
	return r.execute(ctx, graph, schedule, opts, false)
}

// Resume loads a paused graph and continues from the last checkpoint (spec §15, §5.5).
func (r *DefaultRunner) Resume(ctx context.Context, graphID string, opts RunOptions) (GraphRunResult, error) {
	graph, err := r.Repo.Load(graphID)
	if err != nil {
		return GraphRunResult{}, err
	}

	switch graph.Status {
	case GraphStatusPaused:
		if err := TransitionGraph(graph.Status, GraphStatusRunning, false); err != nil {
			return GraphRunResult{}, err
		}
		graph.Status = GraphStatusRunning
	case GraphStatusReady, GraphStatusBlocked, GraphStatusFailed:
		if err := TransitionGraph(graph.Status, GraphStatusRunning, false); err != nil {
			return GraphRunResult{}, err
		}
		graph.Status = GraphStatusRunning
	case GraphStatusRunning:
		// already running
	default:
		return GraphRunResult{}, fmt.Errorf("graph %s: cannot resume from status %q", graphID, graph.Status)
	}

	checkpoint, hasCheckpoint, err := r.Repo.LoadLatestCheckpoint(graphID)
	if err != nil {
		return GraphRunResult{}, err
	}
	if hasCheckpoint {
		graph = markNodesThroughCheckpoint(graph, checkpoint.AfterNode)
	}

	if !hasCheckpoint {
		if _, _, err := r.Repo.Save(graph); err != nil {
			return GraphRunResult{}, err
		}
		r.emit(runtime.EventGraphStarted, graph, "", map[string]any{"resumed": true})
		return GraphRunResult{GraphID: graph.ID, Status: graph.Status}, nil
	}

	sched := r.scheduler()
	if opts.Scheduler != nil {
		sched = opts.Scheduler
	}
	schedule, err := sched.Schedule(ctx, ScheduleRequest{Graph: graph, CIMode: opts.CIMode})
	if err != nil {
		return GraphRunResult{}, err
	}

	return r.execute(ctx, graph, schedule, opts, true)
}

func (r *DefaultRunner) execute(ctx context.Context, graph ExecutionGraph, schedule ExecutionSchedule, opts RunOptions, resumed bool) (GraphRunResult, error) {
	_ = ctx
	strict := opts.StrictTrust

	if graph.Status == GraphStatusPlanned {
		if err := TransitionGraph(graph.Status, GraphStatusReady, false); err != nil {
			return GraphRunResult{}, err
		}
		graph.Status = GraphStatusReady
	}
	if graph.Status == GraphStatusReady {
		if err := TransitionGraph(graph.Status, GraphStatusRunning, false); err != nil {
			return GraphRunResult{}, err
		}
		graph.Status = GraphStatusRunning
	}

	if !resumed {
		r.emit(runtime.EventGraphCreated, graph, "", map[string]any{"status": graph.Status})
	}
	r.emit(runtime.EventGraphStarted, graph, "", map[string]any{"resumed": resumed})

	var costConsumed float64
	started := time.Now()

nodeLoop:
	for _, group := range schedule.ParallelGroups {
		for _, nodeID := range group {
			idx := nodeIndex(graph.Nodes, nodeID)
			if idx < 0 {
				continue
			}
			node := graph.Nodes[idx]
			if IsTerminalNodeStatus(node.Status) {
				continue
			}

			if err := r.transitionNode(&graph, idx, NodeStatusReady); err != nil {
				return GraphRunResult{}, err
			}
			if err := r.transitionNode(&graph, idx, NodeStatusRunning); err != nil {
				return GraphRunResult{}, err
			}
			node = graph.Nodes[idx]

			r.emit(runtime.EventGraphNodeStarted, graph, node.ID, map[string]any{
				"node_type": string(node.Type),
				"agent":     node.Agent,
			})

			if blocked, reason := r.evaluateTrustGate(node, strict, opts); blocked {
				if err := r.transitionNode(&graph, idx, NodeStatusFailed); err != nil {
					return GraphRunResult{}, err
				}
				r.emit(runtime.EventGraphNodeFailed, graph, node.ID, map[string]any{"reason": reason})
				if err := TransitionGraph(graph.Status, GraphStatusBlocked, false); err != nil {
					return GraphRunResult{}, err
				}
				graph.Status = GraphStatusBlocked
				r.emit(runtime.EventGraphBlocked, graph, node.ID, map[string]any{"reason": reason})
				if _, _, err := r.Repo.Save(graph); err != nil {
					return GraphRunResult{}, err
				}
				if strict {
					return GraphRunResult{GraphID: graph.ID, Status: graph.Status}, fmt.Errorf("%w: %s", ErrTrustGateBlocked, reason)
				}
				break nodeLoop
			}

			if err := r.transitionNode(&graph, idx, NodeStatusSucceeded); err != nil {
				return GraphRunResult{}, err
			}
			node = graph.Nodes[idx]
			costConsumed += node.EstimatedCost

			r.emit(runtime.EventGraphNodeCompleted, graph, node.ID, map[string]any{
				"estimated_cost": node.EstimatedCost,
			})

			if shouldPersistCheckpoint(graph, node.ID) {
				gitRef, gitDirty := CaptureGitState(r.RepoRoot)
				state := CheckpointState{
					AfterNode:    node.ID,
					GitRef:       gitRef,
					GitDirty:     gitDirty,
					Outputs:      append([]string(nil), node.Outputs...),
					Validations:  append([]string(nil), node.RequiredChecks...),
					CostConsumed: costConsumed,
					Duration:     time.Since(started).Truncate(time.Second).String(),
				}
				path, err := r.Repo.SaveCheckpoint(graph.ID, state)
				if err != nil {
					return GraphRunResult{}, err
				}
				r.emit(runtime.EventGraphCheckpointCreated, graph, node.ID, map[string]any{
					"checkpoint_path": path,
					"git_ref":         gitRef,
					"cost_consumed":   costConsumed,
				})
			}
		}
	}

	if graph.Status == GraphStatusRunning {
		if err := TransitionGraph(graph.Status, GraphStatusCompleted, false); err != nil {
			return GraphRunResult{}, err
		}
		graph.Status = GraphStatusCompleted
		r.emit(runtime.EventGraphCompleted, graph, "", map[string]any{
			"cost_consumed": costConsumed,
			"duration":      time.Since(started).Truncate(time.Second).String(),
		})
	}

	if _, _, err := r.Repo.Save(graph); err != nil {
		return GraphRunResult{}, err
	}

	return GraphRunResult{GraphID: graph.ID, Status: graph.Status}, nil
}

func (r *DefaultRunner) evaluateTrustGate(node GraphNode, strict bool, opts RunOptions) (blocked bool, reason string) {
	if node.Type != NodeTypeTrustVerification {
		return false, ""
	}
	if opts.DryRun {
		return false, ""
	}
	if !strict {
		return false, ""
	}
	eval := opts.Gates
	if !eval.Configured() {
		return true, "strict trust enabled but verification gates not configured"
	}
	result := eval.Evaluate(context.Background(), confidence.Report{Overall: 0.9}, nil)
	if result.Status == trust.GateStatusBlocked {
		return true, result.Reason
	}
	return false, ""
}

func (r *DefaultRunner) transitionNode(graph *ExecutionGraph, idx int, to NodeStatus) error {
	from := graph.Nodes[idx].Status
	if from == "" {
		from = NodeStatusPending
		graph.Nodes[idx].Status = from
	}
	if err := TransitionNode(from, to, false); err != nil {
		return err
	}
	graph.Nodes[idx].Status = to
	return nil
}

func (r *DefaultRunner) emit(eventType string, graph ExecutionGraph, nodeID string, payload map[string]any) {
	if payload == nil {
		payload = make(map[string]any)
	}
	if nodeID != "" {
		payload["node_id"] = nodeID
	}
	if r.Events != nil {
		_ = r.Events.Emit(eventType, graph.ID, graph.Flow, payload)
	}
	_ = r.Repo.AppendGraphEvent(graph.ID, eventType, payload)
}

func (r *DefaultRunner) scheduler() GraphScheduler {
	if r == nil {
		return DefaultScheduler{}
	}
	return DefaultScheduler{}
}

func nodeIndex(nodes []GraphNode, id string) int {
	for i := range nodes {
		if nodes[i].ID == id {
			return i
		}
	}
	return -1
}

func shouldPersistCheckpoint(graph ExecutionGraph, nodeID string) bool {
	for _, cp := range graph.Checkpoints {
		if cp.After == nodeID {
			return true
		}
	}
	return false
}

func markNodesThroughCheckpoint(graph ExecutionGraph, afterNodeID string) ExecutionGraph {
	reached := false
	order := topologicalNodeOrder(graph.Nodes, graph.Edges)
	orderSet := make(map[string]int, len(order))
	for i, id := range order {
		orderSet[id] = i
	}
	targetOrder, ok := orderSet[afterNodeID]
	if !ok {
		targetOrder = len(order)
	}

	for i := range graph.Nodes {
		id := graph.Nodes[i].ID
		nodeOrder, known := orderSet[id]
		if !known {
			continue
		}
		if nodeOrder <= targetOrder {
			if graph.Nodes[i].Status == "" || graph.Nodes[i].Status == NodeStatusPending || graph.Nodes[i].Status == NodeStatusReady {
				graph.Nodes[i].Status = NodeStatusSucceeded
			}
			if id == afterNodeID {
				reached = true
			}
		}
	}
	if !reached {
		for i := range graph.Nodes {
			if graph.Nodes[i].ID == afterNodeID {
				graph.Nodes[i].Status = NodeStatusSucceeded
			}
		}
	}
	return graph
}
