package report

import (
	"fmt"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

// FromRunMetric maps persisted run_metrics into a report section (specv3 §15).
func FromRunMetric(m telemetry.RunMetric) *CostPerformance {
	c := &CostPerformance{
		EstimatedInputTokens:  m.EstimatedInputTokens,
		EstimatedOutputTokens: m.EstimatedOutputTokens,
		ActualInputTokens:     m.ActualInputTokens,
		ActualOutputTokens:    m.ActualOutputTokens,
		EstimatedCost:         formatEUR(m.EstimatedCostCents),
		ActualCost:            formatEUR(m.ActualCostCents),
		EstimatedDuration:     m.EstimatedDuration.Round(time.Second).String(),
		ActualDuration:        m.ActualDuration.Round(time.Second).String(),
	}
	return c
}

func formatEUR(cents int64) string {
	if cents < 0 {
		return fmt.Sprintf("-€%d.%02d", (-cents)/100, (-cents)%100)
	}
	return fmt.Sprintf("€%d.%02d", cents/100, cents%100)
}
