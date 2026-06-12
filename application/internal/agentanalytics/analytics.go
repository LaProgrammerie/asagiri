package agentanalytics

import (
	"math"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
)

const ReportVersion = "agents-stats-v1"

// Options filters ledger entries before aggregation.
type Options struct {
	AgentID  string
	Provider string
}

// Report is the analytics read model for `asa agents stats`.
type Report struct {
	ReportVersion string       `json:"report_version"`
	LedgerPath    string       `json:"ledger_path"`
	Filter        Filter       `json:"filter,omitempty"`
	Global        Stats        `json:"global"`
	ByAgent       []GroupStats `json:"by_agent"`
	ByProvider    []GroupStats `json:"by_provider"`
}

// Filter echoes CLI filters applied to the dataset.
type Filter struct {
	AgentID  string `json:"agent_id,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// Stats summarizes one aggregation bucket.
type Stats struct {
	TotalRuns          int      `json:"total_runs"`
	SuccessCount       int      `json:"success_count"`
	FailureCount       int      `json:"failure_count"`
	AvgDurationMS      float64  `json:"avg_duration_ms"`
	P95DurationMS      int64    `json:"p95_duration_ms"`
	ContractChecked    int      `json:"contract_checked"`
	ContractValid      int      `json:"contract_valid"`
	ContractValidRatio *float64 `json:"contract_valid_ratio,omitempty"`
	LastRunAt          string   `json:"last_run_at,omitempty"`
}

// GroupStats is Stats keyed by agent_id or provider.
type GroupStats struct {
	ID string `json:"id"`
	Stats
}

// Build aggregates ledger entries into global, per-agent and per-provider stats.
func Build(repoRoot string, opts Options) (Report, error) {
	ledger, err := agentledger.List(repoRoot, agentledger.ListOptions{})
	if err != nil {
		return Report{}, err
	}

	entries := filterEntries(ledger.Entries, opts)
	report := Report{
		ReportVersion: ReportVersion,
		LedgerPath:    agentledger.LedgerPath(),
		Filter: Filter{
			AgentID:  strings.TrimSpace(opts.AgentID),
			Provider: strings.TrimSpace(opts.Provider),
		},
		Global:     computeStats(entries),
		ByAgent:    groupStats(entries, groupByAgent),
		ByProvider: groupStats(entries, groupByProvider),
	}
	return report, nil
}

func filterEntries(entries []agentledger.Entry, opts Options) []agentledger.Entry {
	agentFilter := strings.TrimSpace(opts.AgentID)
	providerFilter := strings.TrimSpace(opts.Provider)
	if agentFilter == "" && providerFilter == "" {
		return append([]agentledger.Entry(nil), entries...)
	}
	out := make([]agentledger.Entry, 0, len(entries))
	for _, e := range entries {
		if agentFilter != "" && e.AgentID != agentFilter {
			continue
		}
		if providerFilter != "" && strings.TrimSpace(e.Provider) != providerFilter {
			continue
		}
		out = append(out, e)
	}
	return out
}

func groupByAgent(e agentledger.Entry) string {
	return strings.TrimSpace(e.AgentID)
}

func groupByProvider(e agentledger.Entry) string {
	p := strings.TrimSpace(e.Provider)
	if p == "" {
		return "—"
	}
	return p
}

func groupStats(entries []agentledger.Entry, keyFn func(agentledger.Entry) string) []GroupStats {
	buckets := map[string][]agentledger.Entry{}
	for _, e := range entries {
		key := keyFn(e)
		if key == "" {
			key = "—"
		}
		buckets[key] = append(buckets[key], e)
	}
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]GroupStats, 0, len(keys))
	for _, k := range keys {
		out = append(out, GroupStats{ID: k, Stats: computeStats(buckets[k])})
	}
	return out
}

func computeStats(entries []agentledger.Entry) Stats {
	stats := Stats{}
	if len(entries) == 0 {
		return stats
	}
	stats.TotalRuns = len(entries)
	var durationSum int64
	durations := make([]int64, 0, len(entries))
	var lastRun string
	for _, e := range entries {
		if e.ExitCode == 0 {
			stats.SuccessCount++
		} else {
			stats.FailureCount++
		}
		if e.DurationMS > 0 {
			durationSum += e.DurationMS
			durations = append(durations, e.DurationMS)
		}
		if e.ContractValid != nil {
			stats.ContractChecked++
			if *e.ContractValid {
				stats.ContractValid++
			}
		}
		if e.StartedAt > lastRun {
			lastRun = e.StartedAt
		}
	}
	if len(durations) > 0 {
		stats.AvgDurationMS = float64(durationSum) / float64(len(durations))
		stats.P95DurationMS = percentile95(durations)
	}
	stats.LastRunAt = lastRun
	if stats.ContractChecked > 0 {
		ratio := float64(stats.ContractValid) / float64(stats.ContractChecked)
		stats.ContractValidRatio = &ratio
	}
	return stats
}

func percentile95(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(math.Ceil(0.95*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
