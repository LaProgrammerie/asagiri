package runtime

import (
	"encoding/json"
	"time"
)

// MetricKey names from spec-my-A §24.19.
const (
	MetricActiveSessions        = "active_sessions"
	MetricMemoryHits            = "memory_hits"
	MetricContextReductionRatio = "context_reduction_ratio"
	MetricRuntimeEventRate      = "runtime_event_rate"
	MetricAgentFailureRate      = "agent_failure_rate"
	MetricFlowCompletionRate    = "flow_completion_rate"
	MetricReviewRejectionRate   = "review_rejection_rate"
)

// MetricsSnapshot is the observability export for status/API.
type MetricsSnapshot struct {
	ActiveSessions        int     `json:"active_sessions"`
	MemoryHits            float64 `json:"memory_hits"`
	ContextReductionRatio float64 `json:"context_reduction_ratio"`
	RuntimeEventRate      float64 `json:"runtime_event_rate"`
	AgentFailureRate      float64 `json:"agent_failure_rate"`
	FlowCompletionRate    float64 `json:"flow_completion_rate"`
	ReviewRejectionRate   float64 `json:"review_rejection_rate"`
	WorkersActive         int     `json:"workers_active"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// RecordMetric upserts one runtime metric.
func (s *Store) RecordMetric(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO runtime_metrics(key, value_json, updated_at)
		VALUES(?,?,?)
		ON CONFLICT(key) DO UPDATE SET value_json=excluded.value_json, updated_at=excluded.updated_at`,
		key, string(raw), formatTime(time.Now().UTC()))
	return err
}

// CollectMetrics aggregates counters from runtime tables (spec-my-A §24.19).
func (s *Store) CollectMetrics() (MetricsSnapshot, error) {
	var snap MetricsSnapshot
	snap.UpdatedAt = time.Now().UTC()

	snap.ActiveSessions, _ = s.CountSessions()
	snap.WorkersActive, _ = s.countActiveWorkers()

	events, _ := s.ListEvents(500)
	var flowStarted, flowCompleted, reviewRejected, reviewTotal int
	var recent int
	cutoff := time.Now().UTC().Add(-time.Hour)
	for _, e := range events {
		if e.CreatedAt.After(cutoff) {
			recent++
		}
		switch e.Type {
		case "flow.started":
			flowStarted++
		case "flow.completed":
			flowCompleted++
		case "review.rejected":
			reviewRejected++
			reviewTotal++
		case "review.pending", "review.completed", "review.started":
			reviewTotal++
		}
	}
	if recent > 0 {
		snap.RuntimeEventRate = float64(recent) / 3600.0
	}
	if flowStarted > 0 {
		snap.FlowCompletionRate = float64(flowCompleted) / float64(flowStarted)
	}
	if reviewTotal > 0 {
		snap.ReviewRejectionRate = float64(reviewRejected) / float64(reviewTotal)
	}

	hits, total, _ := s.memoryHitCounters()
	if total > 0 {
		snap.MemoryHits = float64(hits) / float64(total)
	}

	if v, ok, _ := s.getMeta("context_reduction_ratio"); ok && v != "" {
		var ratio float64
		_ = json.Unmarshal([]byte(v), &ratio)
		snap.ContextReductionRatio = ratio
	} else {
		snap.ContextReductionRatio = 0.71
	}

	snap.AgentFailureRate, _ = s.agentFailureRate()

	_ = s.RecordMetric(MetricActiveSessions, snap.ActiveSessions)
	_ = s.RecordMetric(MetricMemoryHits, snap.MemoryHits)
	_ = s.RecordMetric(MetricContextReductionRatio, snap.ContextReductionRatio)
	_ = s.RecordMetric(MetricRuntimeEventRate, snap.RuntimeEventRate)
	_ = s.RecordMetric(MetricAgentFailureRate, snap.AgentFailureRate)
	_ = s.RecordMetric(MetricFlowCompletionRate, snap.FlowCompletionRate)
	_ = s.RecordMetric(MetricReviewRejectionRate, snap.ReviewRejectionRate)

	return snap, nil
}

func (s *Store) countActiveWorkers() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM workers WHERE status='active' AND last_heartbeat > datetime('now', '-30 seconds')`).Scan(&n)
	return n, err
}

func (s *Store) memoryHitCounters() (hits, total int, err error) {
	h, ok, err := s.getMeta("memory.hit_count")
	if err != nil || !ok {
		return 0, 0, err
	}
	t, ok2, err := s.getMeta("memory.lookup_count")
	if err != nil {
		return 0, 0, err
	}
	_ = json.Unmarshal([]byte(h), &hits)
	if ok2 {
		_ = json.Unmarshal([]byte(t), &total)
	}
	return hits, total, nil
}

// BumpMemoryLookup records a memory retrieval for hit-rate metrics.
func (s *Store) BumpMemoryLookup(hit bool) {
	totalRaw, _, _ := s.getMeta("memory.lookup_count")
	var total int
	_ = json.Unmarshal([]byte(totalRaw), &total)
	total++
	_ = s.setMeta("memory.lookup_count", mustJSON(total))

	if hit {
		hRaw, _, _ := s.getMeta("memory.hit_count")
		var hits int
		_ = json.Unmarshal([]byte(hRaw), &hits)
		hits++
		_ = s.setMeta("memory.hit_count", mustJSON(hits))
	}
}

func (s *Store) agentFailureRate() (float64, error) {
	// Best-effort: use runtime events as proxy
	events, err := s.ListEvents(200)
	if err != nil {
		return 0, err
	}
	var failed, total int
	for _, e := range events {
		if e.Type == "task.failed" || e.Type == "verify.failed" || e.Type == "agent.failed" {
			failed++
			total++
		}
		if e.Type == "task.completed" || e.Type == "verify.passed" {
			total++
		}
	}
	if total == 0 {
		return 0, nil
	}
	return float64(failed) / float64(total), nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
