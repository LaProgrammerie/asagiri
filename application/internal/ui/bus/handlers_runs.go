package bus

import (
	"context"
	"encoding/json"
	"strings"
)

// rawStep mirrors workflow.StepState (run.steps_json) without importing the
// workflow service into the UI bus.
type rawStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// handleGetRunDetail aggregates one run's detail from the state store
// (run + tasks + metrics), trust summary, active agents and recent events.
// Aggregation only — no business logic, per ADR-027.
func (b *queryBus) handleGetRunDetail(ctx context.Context, q GetRunDetailQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(q.RunID)
	if id == "" {
		return RunDetail{Warning: "no run selected"}, nil
	}

	store, err := b.deps.StateOpen(b.deps.StateDBPath)
	if err != nil {
		return RunDetail{ID: id, Warning: err.Error()}, nil
	}
	defer store.Close()
	if err := store.Migrate(); err != nil {
		return RunDetail{ID: id, Warning: err.Error()}, nil
	}

	run, err := store.GetRun(id)
	if err != nil {
		return RunDetail{ID: id, Warning: err.Error()}, nil
	}
	if run == nil {
		return RunDetail{ID: id, Warning: "run not found"}, nil
	}

	detail := RunDetail{
		ID:        run.ID,
		Feature:   run.Feature,
		Status:    run.Status,
		CreatedAt: run.CreatedAt,
		UpdatedAt: run.UpdatedAt,
	}
	detail.Pipeline = parsePipeline(run.StepsJSON)
	detail.Validation = validationFromPipeline(detail.Pipeline)

	if tasks, terr := store.ListTasksByRun(id); terr == nil {
		for _, t := range tasks {
			if strings.TrimSpace(t.WorktreePath) != "" {
				detail.Worktree = t.WorktreePath
				break
			}
		}
	}

	if metric, merr := store.GetRunMetric(id); merr == nil && metric != nil {
		cents := metric.ActualCostCents
		if cents == 0 {
			cents = metric.EstimatedCostCents
		}
		detail.CostEUR = float64(cents) / 100.0
	}

	// Reuse existing read handlers for cross-cutting data (trust, agents, events).
	if trustAny, terr := b.handleGetTrustSummary(ctx, GetTrustSummaryQuery{}); terr == nil {
		if typed, ok := trustAny.(TrustSummaryResult); ok {
			detail.TrustGate = typed
		}
	}
	if agentsAny, aerr := b.handleListActiveAgents(ctx, ListActiveAgentsQuery{Limit: 50}); aerr == nil {
		if typed, ok := agentsAny.(ActiveAgentsResult); ok {
			detail.Agents = typed.Agents
		}
	}
	if eventsAny, eerr := b.handleGetRecentEvents(ctx, GetRecentEventsQuery{Limit: 8}); eerr == nil {
		if typed, ok := eventsAny.(RecentEventsResult); ok {
			detail.Events = filterRunEvents(typed.Events, run.Feature, run.ID)
		}
	}
	return detail, nil
}

func parsePipeline(stepsJSON string) []RunPipelineStep {
	stepsJSON = strings.TrimSpace(stepsJSON)
	if stepsJSON == "" {
		return nil
	}
	var raw []rawStep
	if err := json.Unmarshal([]byte(stepsJSON), &raw); err != nil {
		return nil
	}
	out := make([]RunPipelineStep, 0, len(raw))
	for _, s := range raw {
		out = append(out, RunPipelineStep{ID: s.Name, Label: s.Name, Status: s.Status})
	}
	return out
}

// validationFromPipeline summarises the verification step state when present.
func validationFromPipeline(steps []RunPipelineStep) string {
	for _, s := range steps {
		name := strings.ToLower(s.ID)
		if strings.Contains(name, "verify") || strings.Contains(name, "valid") ||
			strings.Contains(name, "test") || strings.Contains(name, "trust") {
			return s.ID + ": " + s.Status
		}
	}
	return "—"
}

// filterRunEvents keeps events related to the run by flow/session, falling back
// to the recent events when no per-run correlation is available.
func filterRunEvents(events []EventSummary, feature, runID string) []EventSummary {
	feature = strings.TrimSpace(feature)
	runID = strings.TrimSpace(runID)
	var matched []EventSummary
	for _, ev := range events {
		if (feature != "" && (ev.FlowID == feature || strings.Contains(ev.FlowID, feature))) ||
			(runID != "" && ev.SessionID == runID) {
			matched = append(matched, ev)
		}
	}
	if len(matched) > 0 {
		return matched
	}
	return events
}
