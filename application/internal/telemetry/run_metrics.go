package telemetry

import (
	"context"
	"time"
)

// SaveRunStarted writes or updates a run metric with start metadata.
func SaveRunStarted(ctx context.Context, st MetricsStore, runID, feature, taskID string, startedAt time.Time) error {
	if st == nil {
		return nil
	}
	return st.SaveRunMetric(ctx, RunMetric{
		RunID:     runID,
		Feature:   feature,
		TaskID:    taskID,
		StartedAt: startedAt,
		Status:    "running",
	})
}

// SaveRunFinished patches duration and token totals when a run completes.
func SaveRunFinished(ctx context.Context, st MetricsStore, m RunMetric) error {
	if st == nil {
		return nil
	}
	return st.SaveRunMetric(ctx, m)
}
