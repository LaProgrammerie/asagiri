package memory

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// Engine provides scoped memory retrieval, scoring, aging, and consolidation (spec-my-A §24.10–13).
type Engine struct {
	store *runtime.Store
}

// NewEngine wraps a runtime store.
func NewEngine(store *runtime.Store) *Engine {
	return &Engine{store: store}
}

// Retrieve returns memory entries for a scope, ordered by relevance.
func (e *Engine) Retrieve(scope runtime.MemoryScope, tags []string, limit int) ([]runtime.MemoryEntry, error) {
	if e == nil || e.store == nil {
		return nil, fmt.Errorf("memory: store required")
	}
	entries, err := e.store.ListMemory(scope, limit*3)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return trimLimit(entries, limit), nil
	}
	var filtered []runtime.MemoryEntry
	for _, ent := range entries {
		if tagOverlap(ent.Tags, tags) {
			filtered = append(filtered, ent)
		}
	}
	return trimLimit(filtered, limit), nil
}

// Score adjusts relevance from recency and usage (spec-my-A §24 scoring).
func Score(entry runtime.MemoryEntry, now time.Time) float64 {
	base := entry.Relevance
	if base <= 0 {
		base = 0.5
	}
	age := now.Sub(entry.LastUsedAt)
	if age < 24*time.Hour {
		return clamp01(base * 1.1)
	}
	if age > 30*24*time.Hour {
		return clamp01(base * 0.6)
	}
	return clamp01(base)
}

// Age lowers relevance for entries older than maxAge without deleting.
func (e *Engine) Age(maxAge time.Duration) (int, error) {
	entries, err := e.store.ListMemory("", 0)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	var n int
	for _, ent := range entries {
		if now.Sub(ent.LastUsedAt) <= maxAge {
			continue
		}
		ent.Relevance = clamp01(ent.Relevance * 0.85)
		if err := e.store.UpsertMemory(ent); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// Consolidate merges near-duplicate summaries within the same scope.
func (e *Engine) Consolidate() (int, error) {
	entries, err := e.store.ListMemory("", 0)
	if err != nil {
		return 0, err
	}
	seen := map[string]runtime.MemoryEntry{}
	var merged int
	for _, ent := range entries {
		key := string(ent.Scope) + "|" + normalizeSummary(ent.Summary)
		if prev, ok := seen[key]; ok {
			prev.Relevance = clamp01(prev.Relevance + ent.Relevance*0.25)
			prev.Tags = unionTags(prev.Tags, ent.Tags)
			_ = e.store.UpsertMemory(prev)
			merged++
			continue
		}
		seen[key] = ent
	}
	return merged, nil
}

func trimLimit(in []runtime.MemoryEntry, limit int) []runtime.MemoryEntry {
	if limit <= 0 || len(in) <= limit {
		return in
	}
	return in[:limit]
}

func tagOverlap(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	set := map[string]struct{}{}
	for _, t := range have {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, t := range want {
		if _, ok := set[strings.ToLower(t)]; ok {
			return true
		}
	}
	return false
}

func normalizeSummary(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func unionTags(a, b []string) []string {
	m := map[string]struct{}{}
	var out []string
	for _, t := range append(a, b...) {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := m[t]; ok {
			continue
		}
		m[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
