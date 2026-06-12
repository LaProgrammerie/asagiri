package executiongraph

import (
	"context"
	"fmt"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

// RunOptions configures graph execution (spec §5.3, §19).
type RunOptions struct {
	DryRun          bool
	CIMode          bool
	StrictTrust     bool
	CheckpointEvery string
	Scheduler       GraphScheduler
	Gates           trust.GateEvaluator
	TrustEngine     *trust.Engine
	Coordinator     GraphCoordinator
	NodeExecutor    NodeExecutor
}

// DefaultRunner performs graph execution with runtime events and checkpoints (spec §19–20).
type DefaultRunner struct {
	Repo         *Repository
	RepoRoot     string
	Events       *runtime.GraphEmitter
	Coordinator  GraphCoordinator
	NodeExecutor NodeExecutor
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
func (r *DefaultRunner) DryRun(ctx context.Context, graph ExecutionGraph, ciMode bool, opts ...RunOptions) (GraphRunResult, ExecutionSchedule, GraphArtifacts, error) {
	var runOpts RunOptions
	if len(opts) > 0 {
		runOpts = opts[0]
	}
	graph, _, err := r.coordinate(ctx, graph, runOpts)
	if err != nil {
		return GraphRunResult{}, ExecutionSchedule{}, GraphArtifacts{}, err
	}

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
		result, _, _, err := r.DryRun(ctx, graph, opts.CIMode, opts)
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
		return GraphRunResult{}, fmt.Errorf("%w: graph %s has no checkpoint to resume from", ErrNoCheckpoint, graphID)
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
	strict := opts.StrictTrust

	coordinated := opts.Coordinator != nil
	graph, coordAssignments, err := r.coordinate(ctx, graph, opts)
	if err != nil {
		return GraphRunResult{}, err
	}
	asgByNode := assignmentIndex(coordAssignments)

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
		var lastExecutedNode string
		groupHadExecution := false
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

			nodeExec := opts.NodeExecutor
			if nodeExec == nil {
				nodeExec = r.NodeExecutor
			}
			if nodeExec != nil {
				asg := asgByNode[node.ID]
				if err := nodeExec(ctx, graph, node, asg); err != nil {
					if err := r.transitionNode(&graph, idx, NodeStatusFailed); err != nil {
						return GraphRunResult{}, err
					}
					reason := err.Error()
					r.emit(runtime.EventGraphNodeFailed, graph, node.ID, map[string]any{"reason": reason})
					if coordinated {
						r.emitAgentFailed(graph, node, asgByNode, reason)
					}
					if err := TransitionGraph(graph.Status, GraphStatusFailed, false); err != nil {
						return GraphRunResult{}, err
					}
					graph.Status = GraphStatusFailed
					if _, _, err := r.Repo.Save(graph); err != nil {
						return GraphRunResult{}, err
					}
					return GraphRunResult{GraphID: graph.ID, Status: graph.Status}, err
				}
			}

			if blocked, reason := r.evaluateTrustGate(ctx, graph, node, strict, opts); blocked {
				if err := r.transitionNode(&graph, idx, NodeStatusFailed); err != nil {
					return GraphRunResult{}, err
				}
				r.emit(runtime.EventGraphNodeFailed, graph, node.ID, map[string]any{"reason": reason})
				if coordinated {
					r.emitAgentFailed(graph, node, asgByNode, reason)
				}
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
			if coordinated {
				r.emitAgentCompleted(graph, node, asgByNode)
			}

			lastExecutedNode = node.ID
			groupHadExecution = true

			if ShouldPersistCheckpoint(graph, node.ID, opts.CheckpointEvery) {
				if err := r.persistCheckpoint(graph, node, costConsumed, time.Since(started)); err != nil {
					return GraphRunResult{}, err
				}
			}
		}
		if ShouldPersistGroupCheckpoint(opts.CheckpointEvery) && groupHadExecution && lastExecutedNode != "" {
			idx := nodeIndex(graph.Nodes, lastExecutedNode)
			if idx >= 0 {
				if err := r.persistCheckpoint(graph, graph.Nodes[idx], costConsumed, time.Since(started)); err != nil {
					return GraphRunResult{}, err
				}
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

func (r *DefaultRunner) evaluateTrustGate(ctx context.Context, graph ExecutionGraph, node GraphNode, strict bool, opts RunOptions) (blocked bool, reason string) {
	if node.Type != NodeTypeTrustVerification {
		return false, ""
	}
	if opts.DryRun {
		return false, ""
	}
	if !strict {
		return false, ""
	}

	eng := opts.TrustEngine
	if eng == nil {
		eng = trust.NewEngine(r.RepoRoot)
		if opts.Gates.Configured() {
			eng.Gates = opts.Gates
		}
	}

	verifyResult, err := eng.Verify(ctx, trust.VerificationRequest{
		Flow:    graph.Flow,
		Product: graph.Product,
		Strict:  true,
	})
	if err != nil {
		return true, fmt.Sprintf("trust verification failed: %v", err)
	}
	if verifyResult.Report.Gate.Status == trust.GateStatusNotConfigured && !opts.Gates.Configured() {
		return true, "strict trust enabled but verification gates not configured"
	}
	if verifyResult.Report.Gate.Status == trust.GateStatusBlocked {
		return true, verifyResult.Report.Gate.Reason
	}
	if trust.CIShouldFail(verifyResult.Report, true) {
		reason := verifyResult.Report.Gate.Reason
		if reason == "" {
			reason = "trust checks failed under strict mode"
		}
		return true, reason
	}
	return false, ""
}

func (r *DefaultRunner) persistCheckpoint(graph ExecutionGraph, node GraphNode, costConsumed float64, elapsed time.Duration) error {
	gitRef, gitDirty := CaptureGitState(r.RepoRoot)
	state := CheckpointState{
		AfterNode:    node.ID,
		GitRef:       gitRef,
		GitDirty:     gitDirty,
		Outputs:      append([]string(nil), node.Outputs...),
		Validations:  append([]string(nil), node.RequiredChecks...),
		CostConsumed: costConsumed,
		Duration:     elapsed.Truncate(time.Second).String(),
	}
	path, err := r.Repo.SaveCheckpoint(graph.ID, state)
	if err != nil {
		return err
	}
	r.emit(runtime.EventGraphCheckpointCreated, graph, node.ID, map[string]any{
		"checkpoint_path": path,
		"git_ref":         gitRef,
		"cost_consumed":   costConsumed,
	})
	return nil
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

func (r *DefaultRunner) coordinate(ctx context.Context, graph ExecutionGraph, opts RunOptions) (ExecutionGraph, []CoordinationAssignment, error) {
	coord := opts.Coordinator
	if coord == nil && r != nil {
		coord = r.Coordinator
	}
	if coord == nil {
		return graph, nil, nil
	}
	result, err := coord(ctx, graph)
	if err != nil {
		return graph, nil, err
	}
	return result.Graph, result.Assignments, nil
}

func assignmentIndex(assignments []CoordinationAssignment) map[string]CoordinationAssignment {
	out := make(map[string]CoordinationAssignment, len(assignments))
	for _, a := range assignments {
		out[a.NodeID] = a
	}
	return out
}

func (r *DefaultRunner) emitAgentCompleted(graph ExecutionGraph, node GraphNode, asgByNode map[string]CoordinationAssignment) {
	if len(asgByNode) == 0 || r.Events == nil {
		return
	}
	asg, ok := asgByNode[node.ID]
	if !ok {
		asg = CoordinationAssignment{
			NodeID:   node.ID,
			AgentRef: node.Agent,
			Role:     RoleForNodeType(node.Type),
		}
	}
	emitter := &runtime.AgentEmitter{Store: r.Events.Store}
	_ = emitter.Emit(runtime.EventAgentCompleted, graph.ID, graph.Flow, agentAssignmentPayload(asg))
}

func (r *DefaultRunner) emitAgentFailed(graph ExecutionGraph, node GraphNode, asgByNode map[string]CoordinationAssignment, reason string) {
	if len(asgByNode) == 0 || r.Events == nil {
		return
	}
	asg, ok := asgByNode[node.ID]
	if !ok {
		asg = CoordinationAssignment{
			NodeID:   node.ID,
			AgentRef: node.Agent,
			Role:     RoleForNodeType(node.Type),
		}
	}
	payload := agentAssignmentPayload(asg)
	payload["reason"] = reason
	emitter := &runtime.AgentEmitter{Store: r.Events.Store}
	_ = emitter.Emit(runtime.EventAgentFailed, graph.ID, graph.Flow, payload)
}

func agentAssignmentPayload(asg CoordinationAssignment) map[string]any {
	return map[string]any{
		"node_id":    asg.NodeID,
		"agent_ref":  asg.AgentRef,
		"role":       asg.Role,
		"isolation":  asg.Isolation,
		"profile_id": asg.ProfileID,
	}
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
