package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

func durationToMs(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int(d.Milliseconds())
}

func msToDuration(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

func coalesceInt(prev, next int) int {
	if next != 0 {
		return next
	}
	return prev
}

func coalesceInt64(prev, next int64) int64 {
	if next != 0 {
		return next
	}
	return prev
}

func coalesceDur(prev, next time.Duration) time.Duration {
	if next != 0 {
		return next
	}
	return prev
}

func coalesceStr(prev, next string) string {
	if strings.TrimSpace(next) != "" {
		return next
	}
	return prev
}

func mergeRunMetrics(prev *telemetry.RunMetric, next telemetry.RunMetric) telemetry.RunMetric {
	if prev == nil {
		return next
	}
	out := *prev
	out.RunID = next.RunID
	out.Feature = coalesceStr(prev.Feature, next.Feature)
	out.TaskID = coalesceStr(prev.TaskID, next.TaskID)
	if !next.StartedAt.IsZero() {
		out.StartedAt = next.StartedAt
	}
	if !next.FinishedAt.IsZero() {
		out.FinishedAt = next.FinishedAt
	}
	out.EstimatedInputTokens = coalesceInt(prev.EstimatedInputTokens, next.EstimatedInputTokens)
	out.EstimatedOutputTokens = coalesceInt(prev.EstimatedOutputTokens, next.EstimatedOutputTokens)
	out.ActualInputTokens = coalesceInt(prev.ActualInputTokens, next.ActualInputTokens)
	out.ActualOutputTokens = coalesceInt(prev.ActualOutputTokens, next.ActualOutputTokens)
	out.EstimatedCostCents = coalesceInt64(prev.EstimatedCostCents, next.EstimatedCostCents)
	out.ActualCostCents = coalesceInt64(prev.ActualCostCents, next.ActualCostCents)
	out.EstimatedDuration = coalesceDur(prev.EstimatedDuration, next.EstimatedDuration)
	out.ActualDuration = coalesceDur(prev.ActualDuration, next.ActualDuration)
	out.Status = coalesceStr(prev.Status, next.Status)
	return out
}

func (s *Store) getRunMetric(runID string) (*telemetry.RunMetric, error) {
	row := s.db.QueryRow(
		`SELECT run_id, feature, task_id, started_at, finished_at,
			estimated_input_tokens, estimated_output_tokens,
			actual_input_tokens, actual_output_tokens,
			estimated_cost_cents, actual_cost_cents,
			estimated_duration_ms, actual_duration_ms, status
		 FROM run_metrics WHERE run_id = ?`,
		runID,
	)
	var m telemetry.RunMetric
	var feat, task, status sql.NullString
	var startedRaw, finishedRaw sql.NullString
	var estIn, estOut, actIn, actOut sql.NullInt64
	var estCost, actCost sql.NullInt64
	var estMs, actMs sql.NullInt64
	if err := row.Scan(
		&m.RunID, &feat, &task, &startedRaw, &finishedRaw,
		&estIn, &estOut, &actIn, &actOut,
		&estCost, &actCost, &estMs, &actMs, &status,
	); err != nil {
		return nil, err
	}
	m.Feature = feat.String
	m.TaskID = task.String
	m.Status = status.String
	if startedRaw.Valid {
		t, err := parseDBTime(startedRaw.String)
		if err != nil {
			return nil, err
		}
		m.StartedAt = t
	}
	if finishedRaw.Valid {
		t, err := parseDBTime(finishedRaw.String)
		if err != nil {
			return nil, err
		}
		m.FinishedAt = t
	}
	m.EstimatedInputTokens = int(estIn.Int64)
	m.EstimatedOutputTokens = int(estOut.Int64)
	m.ActualInputTokens = int(actIn.Int64)
	m.ActualOutputTokens = int(actOut.Int64)
	m.EstimatedCostCents = estCost.Int64
	m.ActualCostCents = actCost.Int64
	m.EstimatedDuration = msToDuration(int(estMs.Int64))
	m.ActualDuration = msToDuration(int(actMs.Int64))
	return &m, nil
}

// SaveRunMetric inserts or replaces a run_metrics row.
func (s *Store) SaveRunMetric(ctx context.Context, m telemetry.RunMetric) error {
	_ = ctx
	if m.RunID == "" {
		return fmt.Errorf("save run metric: run_id requis")
	}
	var merged telemetry.RunMetric
	if prev, err := s.getRunMetric(m.RunID); err == nil {
		merged = mergeRunMetrics(prev, m)
	} else if errors.Is(err, sql.ErrNoRows) {
		merged = m
	} else {
		return fmt.Errorf("save run metric: %w", err)
	}
	m = merged
	var started, finished *string
	if !m.StartedAt.IsZero() {
		s := timeToText(m.StartedAt)
		started = &s
	}
	if !m.FinishedAt.IsZero() {
		s := timeToText(m.FinishedAt)
		finished = &s
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO run_metrics (
			run_id, feature, task_id, started_at, finished_at,
			estimated_input_tokens, estimated_output_tokens,
			actual_input_tokens, actual_output_tokens,
			estimated_cost_cents, actual_cost_cents,
			estimated_duration_ms, actual_duration_ms, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.RunID, nullStr(m.Feature), nullStr(m.TaskID), started, finished,
		m.EstimatedInputTokens, m.EstimatedOutputTokens,
		m.ActualInputTokens, m.ActualOutputTokens,
		m.EstimatedCostCents, m.ActualCostCents,
		durationToMs(m.EstimatedDuration), durationToMs(m.ActualDuration), nullStr(m.Status),
	)
	if err != nil {
		return fmt.Errorf("save run metric: %w", err)
	}
	return nil
}

// SaveStepMetric inserts or replaces a step_metrics row.
func (s *Store) SaveStepMetric(ctx context.Context, m telemetry.StepMetric) error {
	_ = ctx
	if m.ID == "" || m.RunID == "" {
		return fmt.Errorf("save step metric: id et run_id requis")
	}
	local := 0
	if m.Local {
		local = 1
	}
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO step_metrics (
			id, run_id, step_name, agent, model, local,
			estimated_input_tokens, estimated_output_tokens,
			actual_input_tokens, actual_output_tokens,
			estimated_cost_cents, actual_cost_cents,
			estimated_duration_ms, actual_duration_ms, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.RunID, m.StepName, nullStr(m.Agent), nullStr(m.Model), local,
		m.EstimatedInputTokens, m.EstimatedOutputTokens,
		m.ActualInputTokens, m.ActualOutputTokens,
		m.EstimatedCostCents, m.ActualCostCents,
		durationToMs(m.EstimatedDuration), durationToMs(m.ActualDuration), nullStr(m.Status),
	)
	if err != nil {
		return fmt.Errorf("save step metric: %w", err)
	}
	return nil
}

// SummarizeStepsSince aggregates step_metrics since a timestamp (specv3 §16.2).
func (s *Store) SummarizeStepsSince(ctx context.Context, since time.Time) (telemetry.StepTotals, error) {
	_ = ctx
	var t telemetry.StepTotals
	sinceStr := timeToText(since)
	row := s.db.QueryRow(
		`SELECT COUNT(*),
		 SUM(CASE WHEN local = 1 THEN 1 ELSE 0 END),
		 SUM(CASE WHEN local = 0 THEN 1 ELSE 0 END)
		 FROM step_metrics
		 WHERE id IN (
		   SELECT sm.id FROM step_metrics sm
		   JOIN run_metrics rm ON sm.run_id = rm.run_id
		   WHERE rm.started_at IS NOT NULL AND rm.started_at >= ?
		 )`,
		sinceStr,
	)
	var local, cloud sql.NullInt64
	if err := row.Scan(&t.StepCount, &local, &cloud); err != nil {
		return t, fmt.Errorf("summarize steps: %w", err)
	}
	t.LocalSteps = int(local.Int64)
	t.CloudSteps = int(cloud.Int64)
	// Token savings: compare estimated input on local vs cloud steps (rough proxy).
	var estLocal, estCloud sql.NullInt64
	_ = s.db.QueryRow(
		`SELECT
		   COALESCE(SUM(CASE WHEN local = 1 THEN estimated_input_tokens ELSE 0 END), 0),
		   COALESCE(SUM(CASE WHEN local = 0 THEN estimated_input_tokens ELSE 0 END), 0)
		 FROM step_metrics sm
		 JOIN run_metrics rm ON sm.run_id = rm.run_id
		 WHERE rm.started_at IS NOT NULL AND rm.started_at >= ?`,
		sinceStr,
	).Scan(&estLocal, &estCloud)
	if estLocal.Int64+estCloud.Int64 > 0 {
		t.AvgTokenSavingsPct = float64(estLocal.Int64) / float64(estLocal.Int64+estCloud.Int64) * 100
	}
	return t, nil
}

// GetRunMetric returns one run_metrics row by id.
func (s *Store) GetRunMetric(runID string) (*telemetry.RunMetric, error) {
	return s.getRunMetric(runID)
}

// QuerySince lists run_metrics with started_at on or after since.
func (s *Store) QuerySince(ctx context.Context, since time.Time) ([]telemetry.RunMetric, error) {
	_ = ctx
	sinceStr := timeToText(since)
	rows, err := s.db.Query(
		`SELECT run_id, feature, task_id, started_at, finished_at,
			estimated_input_tokens, estimated_output_tokens,
			actual_input_tokens, actual_output_tokens,
			estimated_cost_cents, actual_cost_cents,
			estimated_duration_ms, actual_duration_ms, status
		 FROM run_metrics
		 WHERE started_at IS NOT NULL AND started_at >= ?
		 ORDER BY started_at DESC`,
		sinceStr,
	)
	if err != nil {
		return nil, fmt.Errorf("query run metrics: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanRunMetrics(rows)
}

// GetDurationHistory returns recent step durations for modeling.
func (s *Store) GetDurationHistory(ctx context.Context, limit int) ([]telemetry.DurationSample, error) {
	_ = ctx
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT step_name, model, agent, local, actual_duration_ms
		 FROM step_metrics
		 WHERE actual_duration_ms IS NOT NULL AND actual_duration_ms > 0
		 ORDER BY id DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query duration history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []telemetry.DurationSample
	for rows.Next() {
		var step, model, agent string
		var local int
		var ms int
		if err := rows.Scan(&step, &model, &agent, &local, &ms); err != nil {
			return nil, fmt.Errorf("scan duration history: %w", err)
		}
		out = append(out, telemetry.DurationSample{
			StepName: step,
			Model:    model,
			Agent:    agent,
			Local:    local != 0,
			Duration: msToDuration(ms),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func scanRunMetrics(rows *sql.Rows) ([]telemetry.RunMetric, error) {
	var out []telemetry.RunMetric
	for rows.Next() {
		var m telemetry.RunMetric
		var feat, task, status sql.NullString
		var startedRaw, finishedRaw sql.NullString
		var estIn, estOut, actIn, actOut sql.NullInt64
		var estCost, actCost sql.NullInt64
		var estMs, actMs sql.NullInt64
		if err := rows.Scan(
			&m.RunID, &feat, &task, &startedRaw, &finishedRaw,
			&estIn, &estOut, &actIn, &actOut,
			&estCost, &actCost, &estMs, &actMs, &status,
		); err != nil {
			return nil, fmt.Errorf("scan run metric: %w", err)
		}
		m.Feature = feat.String
		m.TaskID = task.String
		m.Status = status.String
		if startedRaw.Valid {
			t, err := parseDBTime(startedRaw.String)
			if err != nil {
				return nil, err
			}
			m.StartedAt = t
		}
		if finishedRaw.Valid {
			t, err := parseDBTime(finishedRaw.String)
			if err != nil {
				return nil, err
			}
			m.FinishedAt = t
		}
		m.EstimatedInputTokens = int(estIn.Int64)
		m.EstimatedOutputTokens = int(estOut.Int64)
		m.ActualInputTokens = int(actIn.Int64)
		m.ActualOutputTokens = int(actOut.Int64)
		m.EstimatedCostCents = estCost.Int64
		m.ActualCostCents = actCost.Int64
		m.EstimatedDuration = msToDuration(int(estMs.Int64))
		m.ActualDuration = msToDuration(int(actMs.Int64))
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// DurationSamples returns recent actual durations for a step/model pair.
func (s *Store) DurationSamples(stepName, model string, limit int) ([]time.Duration, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT actual_duration_ms FROM step_metrics
		 WHERE actual_duration_ms > 0 AND step_name = ? AND model = ?
		 ORDER BY id DESC LIMIT ?`,
		stepName, model, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("duration samples: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []time.Duration
	for rows.Next() {
		var ms int
		if err := rows.Scan(&ms); err != nil {
			return nil, err
		}
		out = append(out, msToDuration(ms))
	}
	return out, rows.Err()
}

// Compile-time check: Store implements telemetry.MetricsStore.
var _ telemetry.MetricsStore = (*Store)(nil)
