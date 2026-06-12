package cost_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/stretchr/testify/require"
)

func TestScoreStrategy_NoData(t *testing.T) {
	s := cost.ScoreStrategy(telemetry.StepTokenTotals{})
	require.Equal(t, "?", s.Grade)
	require.Equal(t, 0.0, s.LocalPct)
}

func TestScoreStrategy_GradeA(t *testing.T) {
	s := cost.ScoreStrategy(tokens(800_000, 80_000, 100_000, 10_000))
	require.Equal(t, "A", s.Grade)
	require.Greater(t, s.LocalPct, 70.0)
}

func TestScoreStrategy_GradeB(t *testing.T) {
	s := cost.ScoreStrategy(tokens(600_000, 60_000, 400_000, 40_000))
	require.Equal(t, "B", s.Grade)
	require.InDelta(t, 60.0, s.LocalPct, 0.5)
}

func TestScoreStrategy_GradeC(t *testing.T) {
	s := cost.ScoreStrategy(tokens(400_000, 40_000, 600_000, 60_000))
	require.Equal(t, "C", s.Grade)
}

func TestScoreStrategy_GradeD(t *testing.T) {
	s := cost.ScoreStrategy(tokens(100_000, 10_000, 900_000, 90_000))
	require.Equal(t, "D", s.Grade)
}

func TestComputeEscalations_NoSteps(t *testing.T) {
	m := cost.ComputeEscalations(telemetry.StepTotals{})
	require.Equal(t, 0, m.TotalSteps)
	require.Equal(t, 0.0, m.EscalationRate)
}

func TestComputeEscalations(t *testing.T) {
	m := cost.ComputeEscalations(telemetry.StepTotals{
		StepCount:  10,
		LocalSteps: 8,
		CloudSteps: 2,
	})
	require.Equal(t, 10, m.TotalSteps)
	require.Equal(t, 8, m.LocalSteps)
	require.Equal(t, 2, m.PremiumEscalations)
	require.InDelta(t, 0.2, m.EscalationRate, 0.001)
}

func TestBuildWindowReport_NoBaseline(t *testing.T) {
	cfg := &config.Config{}
	cfg.Pricing.Currency = "EUR"
	tot := telemetry.CostTotals{RunCount: 5, ActualCostCents: 100}
	tok := tokens(500_000, 50_000, 100_000, 10_000)
	steps := telemetry.StepTotals{StepCount: 10, LocalSteps: 8, CloudSteps: 2}

	w := cost.BuildWindowReport("7d", tot, tok, steps, cfg)

	require.Equal(t, "7d", w.Label)
	require.Equal(t, 5, w.RunCount)
	require.Equal(t, int64(20), w.AvgCostPerRunCents) // 100/5
	require.Equal(t, "A", w.Strategy.Grade)
	require.Equal(t, 2, w.Escalations.PremiumEscalations)
	require.False(t, w.Savings.HasPremiumBaseline())
}

func TestBuildWindowReport_WithBaseline(t *testing.T) {
	cfg := &config.Config{}
	cfg.Pricing.Currency = "EUR"
	cfg.Pricing.PremiumReferenceModel = "gpt-4o"
	cfg.Pricing.Models = map[string]config.ModelPricing{
		"gpt-4o": {InputPer1MTokens: 5.0, OutputPer1MTokens: 15.0},
	}
	tot := telemetry.CostTotals{RunCount: 1, ActualCostCents: 10}
	tok := tokens(1_000_000, 0, 0, 0) // all local

	w := cost.BuildWindowReport("7d", tot, tok, telemetry.StepTotals{}, cfg)

	require.True(t, w.Savings.HasPremiumBaseline())
	require.Equal(t, int64(500), w.Savings.PremiumEquivCents) // 1M × €5/1M = €5.00
	require.Equal(t, int64(490), w.Savings.SavingsCents)
}
