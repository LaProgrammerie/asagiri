package cost

import (
	"context"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// RunHistory feeds duration modeling from prior executions.
type RunHistory struct {
	RecentDurations []time.Duration
}

// DurationHistoryReader optionally supplies SQLite-backed samples.
type DurationHistoryReader interface {
	DurationsFor(stepName, model string, limit int) []time.Duration
}

// DurationModel predicts step latency (specv3 §7).
type DurationModel interface {
	Estimate(ctx context.Context, step PlannedStep, history RunHistory) time.Duration
}

// DefaultDurationModel uses config hints and optional reader heuristics.
type DefaultDurationModel struct {
	Cfg    *config.Config
	Reader DurationHistoryReader
}

func profileLatency(cfg *config.Config, model string) int {
	if cfg == nil {
		return 120
	}
	for _, p := range cfg.Models {
		if p.Model == model && p.TypicalLatencyMsPer1KTokens > 0 {
			return p.TypicalLatencyMsPer1KTokens
		}
	}
	for _, p := range cfg.Models {
		if p.TypicalLatencyMsPer1KTokens > 0 {
			return p.TypicalLatencyMsPer1KTokens
		}
	}
	return 120
}

// Estimate implements DurationModel.
func (d DefaultDurationModel) Estimate(ctx context.Context, step PlannedStep, history RunHistory) time.Duration {
	_ = ctx
	if len(history.RecentDurations) > 0 {
		var sum time.Duration
		for _, s := range history.RecentDurations {
			sum += s
		}
		return sum / time.Duration(len(history.RecentDurations))
	}
	if d.Reader != nil {
		s := d.Reader.DurationsFor(step.Name, step.Model, 12)
		if len(s) > 0 {
			var sum time.Duration
			for _, x := range s {
				sum += x
			}
			return sum / time.Duration(len(s))
		}
	}
	latPer1k := profileLatency(d.Cfg, step.Model)
	tokens := step.InputTokens
	if tokens < 1000 {
		tokens = 1000
	}
	ms := latPer1k * (tokens / 1000)
	if ms < latPer1k {
		ms = latPer1k
	}
	return time.Duration(ms) * time.Millisecond
}
