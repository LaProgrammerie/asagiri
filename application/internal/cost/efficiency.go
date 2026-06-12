package cost

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

// AgentStrategyScore grades local-first routing quality.
//
// Formula (documented, deterministic):
//
//	localPct = (localInputTokens + localOutputTokens) / totalTokens × 100
//	A  ≥ 70%  — strong local-first
//	B  50–70% — balanced, room to improve
//	C  30–50% — cloud-heavy
//	D  < 30%  — almost all premium
//
// Score is purely informational; no ML or opaque heuristic.
type AgentStrategyScore struct {
	Grade       string  // "A", "B", "C", "D", or "?" when no data
	LocalPct    float64 // 0..100
	Description string
}

// ScoreStrategy computes the AgentStrategyScore from token totals.
// Returns Grade "?" with LocalPct 0 when no tokens have been recorded.
func ScoreStrategy(tokens telemetry.StepTokenTotals) AgentStrategyScore {
	total := tokens.LocalInputTokens + tokens.LocalOutputTokens +
		tokens.CloudInputTokens + tokens.CloudOutputTokens
	if total == 0 {
		return AgentStrategyScore{Grade: "?", Description: "no data yet"}
	}
	localTok := tokens.LocalInputTokens + tokens.LocalOutputTokens
	pct := float64(localTok) / float64(total) * 100

	switch {
	case pct >= 70:
		return AgentStrategyScore{Grade: "A", LocalPct: pct, Description: "strong local-first routing"}
	case pct >= 50:
		return AgentStrategyScore{Grade: "B", LocalPct: pct, Description: "balanced, room to improve"}
	case pct >= 30:
		return AgentStrategyScore{Grade: "C", LocalPct: pct, Description: "cloud-heavy routing"}
	default:
		return AgentStrategyScore{Grade: "D", LocalPct: pct, Description: "almost all premium models"}
	}
}

// EscalationMetrics measures premium model usage.
//
// Hypothesis: a "premium escalation" is any step that ran on a cloud model.
// LocalSteps are "escalations avoided". Escalation rate = CloudSteps / TotalSteps.
// All values are derived from step_metrics.local column — no inference.
type EscalationMetrics struct {
	TotalSteps         int
	LocalSteps         int     // escalations avoided
	PremiumEscalations int     // cloud steps
	EscalationRate     float64 // 0..1; 0 when no steps
}

// ComputeEscalations derives escalation metrics from step totals.
func ComputeEscalations(steps telemetry.StepTotals) EscalationMetrics {
	m := EscalationMetrics{
		TotalSteps:         steps.StepCount,
		LocalSteps:         steps.LocalSteps,
		PremiumEscalations: steps.CloudSteps,
	}
	if steps.StepCount > 0 {
		m.EscalationRate = float64(steps.CloudSteps) / float64(steps.StepCount)
	}
	return m
}

// WindowReport captures efficiency metrics for one time window.
type WindowReport struct {
	Label              string // e.g. "30d" or "7d"
	RunCount           int
	ActualCostCents    int64
	AvgCostPerRunCents int64 // 0 when RunCount == 0
	Savings            SavingsReport
	Strategy           AgentStrategyScore
	Escalations        EscalationMetrics
}

// EfficiencyTrends holds two consecutive windows for trend display.
type EfficiencyTrends struct {
	Previous WindowReport // older window (e.g. day 30→15)
	Current  WindowReport // recent window (e.g. day 15→0)
}

// BuildWindowReport constructs a WindowReport from pre-fetched aggregates.
func BuildWindowReport(label string, tot telemetry.CostTotals, tokens telemetry.StepTokenTotals, steps telemetry.StepTotals, cfg *config.Config) WindowReport {
	sav := ComputeSavings(tokens, tot.ActualCostCents, cfg)
	avg := int64(0)
	if tot.RunCount > 0 {
		avg = tot.ActualCostCents / int64(tot.RunCount)
	}
	return WindowReport{
		Label:              label,
		RunCount:           tot.RunCount,
		ActualCostCents:    tot.ActualCostCents,
		AvgCostPerRunCents: avg,
		Savings:            sav,
		Strategy:           ScoreStrategy(tokens),
		Escalations:        ComputeEscalations(steps),
	}
}
