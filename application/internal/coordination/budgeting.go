package coordination

import "context"

// AgentBudget tracks spend for one agent role in a pipeline (spec-my-D §13).
type AgentBudget struct {
	Role       AgentRole `json:"role"`
	CostEUR    float64   `json:"cost_eur"`
	Tokens     int       `json:"tokens"`
	DurationMs int64     `json:"duration_ms"`
	Retries    int       `json:"retries"`
}

// PipelineBudget aggregates agent budgets for a graph run.
type PipelineBudget struct {
	GraphID string        `json:"graph_id"`
	Agents  []AgentBudget `json:"agents"`
	TotalEUR float64      `json:"total_eur"`
}

// BudgetTracker records agent and pipeline spend (Lot 4 stub).
type BudgetTracker interface {
	Record(ctx context.Context, graphID string, asg AgentAssignment, costEUR float64, tokens int, durationMs int64) error
	Summary(ctx context.Context, graphID string) (PipelineBudget, error)
}

// StubBudgetTracker is a Lot-1 placeholder.
type StubBudgetTracker struct{}

func (StubBudgetTracker) Record(_ context.Context, _ string, _ AgentAssignment, _ float64, _ int, _ int64) error {
	return ErrNotImplemented
}

func (StubBudgetTracker) Summary(_ context.Context, _ string) (PipelineBudget, error) {
	return PipelineBudget{}, ErrNotImplemented
}
