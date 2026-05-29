package coordination

import (
	"context"
	"sync"
)

// AgentBudget tracks spend for one agent role in a pipeline (spec-my-D §13).
type AgentBudget struct {
	Role       AgentRole `json:"role"`
	AgentRef   string    `json:"agent_ref,omitempty"`
	CostEUR    float64   `json:"cost_eur"`
	Tokens     int       `json:"tokens"`
	DurationMs int64     `json:"duration_ms"`
	Retries    int       `json:"retries"`
}

// PipelineBudget aggregates agent budgets for a graph run.
type PipelineBudget struct {
	GraphID  string        `json:"graph_id"`
	Agents   []AgentBudget `json:"agents"`
	TotalEUR float64       `json:"total_eur"`
}

// BudgetTracker records agent and pipeline spend.
type BudgetTracker interface {
	Record(ctx context.Context, graphID string, asg AgentAssignment, costEUR float64, tokens int, durationMs int64) error
	Summary(ctx context.Context, graphID string) (PipelineBudget, error)
}

// MemoryBudgetTracker aggregates budgets in memory per graph ID.
type MemoryBudgetTracker struct {
	mu   sync.Mutex
	byID map[string]map[AgentRole]*AgentBudget
}

// NewMemoryBudgetTracker returns an in-memory budget tracker.
func NewMemoryBudgetTracker() *MemoryBudgetTracker {
	return &MemoryBudgetTracker{byID: make(map[string]map[AgentRole]*AgentBudget)}
}

// Record accumulates spend for the assignment role.
func (t *MemoryBudgetTracker) Record(_ context.Context, graphID string, asg AgentAssignment, costEUR float64, tokens int, durationMs int64) error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.byID == nil {
		t.byID = make(map[string]map[AgentRole]*AgentBudget)
	}
	roles := t.byID[graphID]
	if roles == nil {
		roles = make(map[AgentRole]*AgentBudget)
		t.byID[graphID] = roles
	}
	entry, ok := roles[asg.Role]
	if !ok {
		entry = &AgentBudget{Role: asg.Role, AgentRef: asg.AgentRef}
		roles[asg.Role] = entry
	}
	entry.CostEUR += costEUR
	entry.Tokens += tokens
	entry.DurationMs += durationMs
	if asg.AgentRef != "" {
		entry.AgentRef = asg.AgentRef
	}
	return nil
}

// RecordRetry increments retry count for a role.
func (t *MemoryBudgetTracker) RecordRetry(_ context.Context, graphID string, role AgentRole) error {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	roles := t.byID[graphID]
	if roles == nil {
		roles = make(map[AgentRole]*AgentBudget)
		t.byID[graphID] = roles
	}
	entry, ok := roles[role]
	if !ok {
		entry = &AgentBudget{Role: role}
		roles[role] = entry
	}
	entry.Retries++
	return nil
}

// Summary returns aggregated budgets for a graph run.
func (t *MemoryBudgetTracker) Summary(_ context.Context, graphID string) (PipelineBudget, error) {
	if t == nil {
		return PipelineBudget{GraphID: graphID}, nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	roles := t.byID[graphID]
	out := PipelineBudget{GraphID: graphID}
	for _, b := range roles {
		out.Agents = append(out.Agents, *b)
		out.TotalEUR += b.CostEUR
	}
	return out, nil
}
