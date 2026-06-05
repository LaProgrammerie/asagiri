package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/stretchr/testify/require"
)

// openTestStore opens a migrated in-memory SQLite store for testing.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	require.NoError(t, err)
	require.NoError(t, s.Migrate())
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// seedRun inserts a run_metric with the given started_at and actual_cost_cents.
func seedRun(t *testing.T, s *Store, runID string, startedAt time.Time, actualCents int64) {
	t.Helper()
	require.NoError(t, s.SaveRunMetric(context.Background(), telemetry.RunMetric{
		RunID:           runID,
		StartedAt:       startedAt,
		ActualCostCents: actualCents,
		Status:          "done",
	}))
}

// seedStep inserts a step_metric linked to the given run.
func seedStep(t *testing.T, s *Store, stepID, runID string, local bool, inTokens, outTokens int) {
	t.Helper()
	require.NoError(t, s.SaveStepMetric(context.Background(), telemetry.StepMetric{
		ID:                stepID,
		RunID:             runID,
		StepName:          "dev",
		Local:             local,
		ActualInputTokens: inTokens,
		ActualOutputTokens: outTokens,
	}))
}

// ── Boundary correctness ──────────────────────────────────────────────────────

func TestQueryRunsBetween_LowerBoundInclusive(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC().Truncate(time.Second)
	seedRun(t, s, "r1", now, 100) // started_at == since → must be included

	runs, err := s.QueryRunsBetween(context.Background(), now, now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, runs, 1)
}

func TestQueryRunsBetween_UpperBoundExclusive(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC().Truncate(time.Second)
	seedRun(t, s, "r1", now, 100) // started_at == until → must be excluded

	runs, err := s.QueryRunsBetween(context.Background(), now.Add(-time.Hour), now)
	require.NoError(t, err)
	require.Empty(t, runs)
}

// ── Disjoint windows ──────────────────────────────────────────────────────────

func TestQueryRunsBetween_Disjoint_NoDoubleCount(t *testing.T) {
	s := openTestStore(t)
	t0 := time.Now().UTC().Truncate(time.Second)
	mid := t0.Add(15 * 24 * time.Hour)
	end := t0.Add(30 * 24 * time.Hour)

	seedRun(t, s, "r-prev", t0.Add(7*24*time.Hour), 50)   // in [t0, mid)
	seedRun(t, s, "r-curr", t0.Add(20*24*time.Hour), 80)  // in [mid, end)
	seedRun(t, s, "r-edge", mid, 30)                        // started_at == mid → in [mid, end) only

	prev, err := s.QueryRunsBetween(context.Background(), t0, mid)
	require.NoError(t, err)
	require.Len(t, prev, 1)
	require.Equal(t, "r-prev", prev[0].RunID)

	curr, err := s.QueryRunsBetween(context.Background(), mid, end)
	require.NoError(t, err)
	require.Len(t, curr, 2) // r-curr + r-edge
	ids := map[string]bool{curr[0].RunID: true, curr[1].RunID: true}
	require.True(t, ids["r-curr"])
	require.True(t, ids["r-edge"])
}

// ── Token split per window ────────────────────────────────────────────────────

func TestQueryStepTokensBetween_LocalCloudSplit(t *testing.T) {
	s := openTestStore(t)
	t0 := time.Now().UTC().Truncate(time.Second)
	mid := t0.Add(15 * 24 * time.Hour)
	end := t0.Add(30 * 24 * time.Hour)

	// Previous window: 1 local step, 1 cloud step
	seedRun(t, s, "rp", t0.Add(time.Hour), 0)
	seedStep(t, s, "sp-local", "rp", true, 1000, 200)
	seedStep(t, s, "sp-cloud", "rp", false, 500, 100)

	// Current window: 2 local steps only
	seedRun(t, s, "rc", mid.Add(time.Hour), 0)
	seedStep(t, s, "sc-local1", "rc", true, 800, 160)
	seedStep(t, s, "sc-local2", "rc", true, 400, 80)

	prevTok, err := s.QueryStepTokensBetween(context.Background(), t0, mid)
	require.NoError(t, err)
	require.Equal(t, int64(1000), prevTok.LocalInputTokens)
	require.Equal(t, int64(500), prevTok.CloudInputTokens)

	currTok, err := s.QueryStepTokensBetween(context.Background(), mid, end)
	require.NoError(t, err)
	require.Equal(t, int64(1200), currTok.LocalInputTokens)
	require.Equal(t, int64(0), currTok.CloudInputTokens)
}

// ── Cost per window ───────────────────────────────────────────────────────────

func TestQueryRunsBetween_CostSumPerWindow(t *testing.T) {
	s := openTestStore(t)
	t0 := time.Now().UTC().Truncate(time.Second)
	mid := t0.Add(15 * 24 * time.Hour)
	end := t0.Add(30 * 24 * time.Hour)

	seedRun(t, s, "p1", t0.Add(time.Hour), 50)
	seedRun(t, s, "p2", t0.Add(2*time.Hour), 30)
	seedRun(t, s, "c1", mid.Add(time.Hour), 20)

	prev, _ := s.QueryRunsBetween(context.Background(), t0, mid)
	curr, _ := s.QueryRunsBetween(context.Background(), mid, end)

	prevCents := int64(0)
	for _, r := range prev {
		prevCents += r.ActualCostCents
	}
	currCents := int64(0)
	for _, r := range curr {
		currCents += r.ActualCostCents
	}
	require.Equal(t, int64(80), prevCents)
	require.Equal(t, int64(20), currCents)
}

// ── Empty windows ─────────────────────────────────────────────────────────────

func TestQueryRunsBetween_EmptyWindow(t *testing.T) {
	s := openTestStore(t)
	now := time.Now().UTC()
	// No data seeded
	runs, err := s.QueryRunsBetween(context.Background(), now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	require.Empty(t, runs)

	tokens, err := s.QueryStepTokensBetween(context.Background(), now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	require.Equal(t, int64(0), tokens.LocalInputTokens)
	require.Equal(t, int64(0), tokens.CloudInputTokens)

	steps, err := s.SummarizeStepsBetween(context.Background(), now.Add(-24*time.Hour), now)
	require.NoError(t, err)
	require.Equal(t, 0, steps.StepCount)
}

// ── Single run window ─────────────────────────────────────────────────────────

func TestQueryRunsBetween_SingleRun(t *testing.T) {
	s := openTestStore(t)
	t0 := time.Now().UTC().Truncate(time.Second)
	seedRun(t, s, "only", t0, 42)
	seedStep(t, s, "s1", "only", true, 100, 20)

	runs, err := s.QueryRunsBetween(context.Background(), t0, t0.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, runs, 1)
	require.Equal(t, int64(42), runs[0].ActualCostCents)

	tokens, err := s.QueryStepTokensBetween(context.Background(), t0, t0.Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, int64(100), tokens.LocalInputTokens)
}

// ── SummarizeStepsBetween ─────────────────────────────────────────────────────

func TestSummarizeStepsBetween_LocalCloudCount(t *testing.T) {
	s := openTestStore(t)
	t0 := time.Now().UTC().Truncate(time.Second)
	mid := t0.Add(15 * 24 * time.Hour)

	seedRun(t, s, "r1", t0.Add(time.Hour), 0)
	seedStep(t, s, "s-local", "r1", true, 0, 0)
	seedStep(t, s, "s-cloud", "r1", false, 0, 0)

	// Run after mid — must NOT appear in [t0, mid)
	seedRun(t, s, "r2", mid.Add(time.Hour), 0)
	seedStep(t, s, "s-cloud2", "r2", false, 0, 0)

	totals, err := s.SummarizeStepsBetween(context.Background(), t0, mid)
	require.NoError(t, err)
	require.Equal(t, 2, totals.StepCount)
	require.Equal(t, 1, totals.LocalSteps)
	require.Equal(t, 1, totals.CloudSteps)
}
