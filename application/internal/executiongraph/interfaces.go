package executiongraph

import "context"

// GraphPlanRequest groups inputs for graph planning (spec §21).
type GraphPlanRequest struct {
	Product        string
	Flow           string
	FromProduct    bool
	FromSpec       bool
	IncludeReviews bool
	IncludeDocs    bool
	Estimate       bool
}

// TaskBinding links a graph node to task and flow metadata for inference (spec §10).
type TaskBinding struct {
	NodeID      string
	TaskID      string
	FlowStepID  string
	StepIndex   int
	Action      string
	ContractRef string
	Sensitive   bool
	ScopePaths  []string
}

// DependencyInput feeds dependency inference (spec §10).
type DependencyInput struct {
	Product      string
	Flow         string
	RepoRoot     string
	Nodes        []GraphNode
	TaskBindings []TaskBinding
}

// ScheduleRequest feeds graph scheduling (spec §11).
type ScheduleRequest struct {
	Graph  ExecutionGraph
	CIMode bool // forces max_parallel to 1 regardless of graph strategy
}

// BlockedNode describes a node waiting on predecessors (spec §11, §22 UX).
type BlockedNode struct {
	NodeID  string   `json:"node_id"`
	WaitFor []string `json:"wait_for"`
}

// ExecutionSchedule is the output of graph scheduling (spec §11).
type ExecutionSchedule struct {
	GraphID        string        `json:"graph_id"`
	ParallelGroups [][]string    `json:"parallel_groups"`
	Blocked        []BlockedNode `json:"blocked,omitempty"`
}

// GraphRunResult summarizes graph execution (spec §19).
type GraphRunResult struct {
	GraphID string      `json:"graph_id"`
	Status  GraphStatus `json:"status"`
}

// ExecutionGraphPlanner builds execution graphs from product inputs (spec §21).
type ExecutionGraphPlanner interface {
	Build(ctx context.Context, req GraphPlanRequest) (ExecutionGraph, error)
}

// DependencyInferer infers edges between nodes (spec §10).
type DependencyInferer interface {
	Infer(ctx context.Context, input DependencyInput) ([]GraphEdge, error)
}

// GraphScheduler orders nodes for execution (spec §11).
type GraphScheduler interface {
	Schedule(ctx context.Context, req ScheduleRequest) (ExecutionSchedule, error)
}

// GraphExecutor runs or resumes graph execution (spec §19).
type GraphExecutor interface {
	Run(ctx context.Context, graph ExecutionGraph) (GraphRunResult, error)
	Resume(ctx context.Context, graphID string) (GraphRunResult, error)
}

// StubPlanner is a lot-1 placeholder for ExecutionGraphPlanner.
type StubPlanner struct{}

func (StubPlanner) Build(_ context.Context, _ GraphPlanRequest) (ExecutionGraph, error) {
	return ExecutionGraph{}, ErrNotImplemented
}

// StubDependencyInferer is a lot-1 placeholder for DependencyInferer.
type StubDependencyInferer struct{}

func (StubDependencyInferer) Infer(_ context.Context, _ DependencyInput) ([]GraphEdge, error) {
	return nil, ErrNotImplemented
}

// StubScheduler is a lot-1 placeholder for GraphScheduler.
type StubScheduler struct{}

func (StubScheduler) Schedule(_ context.Context, _ ScheduleRequest) (ExecutionSchedule, error) {
	return ExecutionSchedule{}, ErrNotImplemented
}

// StubExecutor is a lot-1 placeholder for GraphExecutor.
type StubExecutor struct{}

func (StubExecutor) Run(_ context.Context, _ ExecutionGraph) (GraphRunResult, error) {
	return GraphRunResult{}, ErrNotImplemented
}

func (StubExecutor) Resume(_ context.Context, _ string) (GraphRunResult, error) {
	return GraphRunResult{}, ErrNotImplemented
}
