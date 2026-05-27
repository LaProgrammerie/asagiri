package memory_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestEngineRetrieveAndConsolidate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:   runtime.ScopeProject,
		Type:    "decision",
		Summary: "use sqlite for runtime",
		Relevance: 0.8,
		Tags:    []string{"architecture"},
	}))
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:   runtime.ScopeProject,
		Type:    "decision",
		Summary: "use sqlite for runtime",
		Relevance: 0.6,
	}))

	eng := memory.NewEngine(store)
	n, err := eng.Consolidate()
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)

	got, err := eng.Retrieve(runtime.ScopeProject, []string{"architecture"}, 10)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	score := memory.Score(got[0], time.Now())
	require.Greater(t, score, 0.0)
}

func TestAge(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	old := time.Now().UTC().Add(-60 * 24 * time.Hour)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope: runtime.ScopeProject, Type: "note", Summary: "old entry",
		Relevance: 0.9, LastUsedAt: old, CreatedAt: old,
	}))
	n, err := memory.NewEngine(store).Age(7 * 24 * time.Hour)
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	_ = filepath.Join(dir, ".asagiri", "runtime", "runtime.db")
}
