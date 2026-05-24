package telemetry

import (
	"context"
	"fmt"
	"time"
)

// CostTotals aggregates historical runs since a timestamp.
type CostTotals struct {
	RunCount            int
	EstimatedCostCents  int64
	ActualCostCents     int64
	EstimatedInputTok   int64
	EstimatedOutputTok  int64
}

// SummarizeSince aggregates run_metrics rows for reporting.
func SummarizeSince(ctx context.Context, st MetricsStore, since time.Time) (CostTotals, error) {
	var zero CostTotals
	if st == nil {
		return zero, nil
	}
	runs, err := st.QuerySince(ctx, since)
	if err != nil {
		return zero, err
	}
	var t CostTotals
	for _, r := range runs {
		t.RunCount++
		t.EstimatedCostCents += r.EstimatedCostCents
		t.ActualCostCents += r.ActualCostCents
		t.EstimatedInputTok += int64(r.EstimatedInputTokens)
		t.EstimatedOutputTok += int64(r.EstimatedOutputTokens)
	}
	return t, nil
}

// RunSummary is formatted cost report for CLI.
type RunSummary struct {
	RunCount      int
	EstimatedCost string
	ActualCost    string
}

// SummarizeRuns builds a display summary from run metrics.
func SummarizeRuns(runs []RunMetric) RunSummary {
	var t CostTotals
	for _, r := range runs {
		t.RunCount++
		t.EstimatedCostCents += r.EstimatedCostCents
		t.ActualCostCents += r.ActualCostCents
	}
	return RunSummary{
		RunCount:      t.RunCount,
		EstimatedCost: formatEUR(t.EstimatedCostCents),
		ActualCost:    formatEUR(t.ActualCostCents),
	}
}

func formatEUR(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return sign + "€" + fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
