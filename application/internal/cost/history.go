package cost

import (
	"context"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

// StoreHistory adapts telemetry.MetricsStore to HistoryReader.
type StoreHistory struct {
	Store telemetry.MetricsStore
}

func (h StoreHistory) AverageDuration(ctx context.Context, agent, model, taskType string) (time.Duration, bool) {
	if h.Store == nil {
		return 0, false
	}
	samples, err := h.Store.GetDurationHistory(ctx, 50)
	if err != nil || len(samples) == 0 {
		return 0, false
	}
	var total time.Duration
	var n int
	for _, s := range samples {
		if agent != "" && s.Agent != agent {
			continue
		}
		if model != "" && s.Model != model {
			continue
		}
		if taskType != "" && s.StepName != taskType {
			continue
		}
		total += s.Duration
		n++
	}
	if n == 0 {
		return 0, false
	}
	return total / time.Duration(n), true
}
