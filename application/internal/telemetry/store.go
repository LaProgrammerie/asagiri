package telemetry

import (
	"context"
	"time"
)

// MetricsStore persists and queries cost/performance metrics.
type MetricsStore interface {
	SaveRunMetric(ctx context.Context, m RunMetric) error
	SaveStepMetric(ctx context.Context, m StepMetric) error
	QuerySince(ctx context.Context, since time.Time) ([]RunMetric, error)
	SummarizeStepsSince(ctx context.Context, since time.Time) (StepTotals, error)
	GetDurationHistory(ctx context.Context, limit int) ([]DurationSample, error)
	QueryStepTokens(ctx context.Context, since time.Time) (StepTokenTotals, error)
	QueryRunsBetween(ctx context.Context, since, until time.Time) ([]RunMetric, error)
	SummarizeStepsBetween(ctx context.Context, since, until time.Time) (StepTotals, error)
	QueryStepTokensBetween(ctx context.Context, since, until time.Time) (StepTokenTotals, error)
}
