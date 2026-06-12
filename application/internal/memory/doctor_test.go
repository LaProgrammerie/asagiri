package memory_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestDoctorHealthyHash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:     runtime.ScopeProject,
		Type:      "note",
		Summary:   "healthy entry",
		Relevance: 0.8,
	}))

	embedder.Configure(embedder.NewHash())
	checks, err := memory.NewEngine(store).Doctor(context.Background())
	require.NoError(t, err)
	require.Len(t, checks, 3)
	for _, c := range checks {
		require.NoError(t, c.Err, c.Name)
	}
}

func TestDoctorDimensionMismatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		ID:            "bad-dim",
		Scope:         runtime.ScopeProject,
		Type:          "note",
		Summary:       "stale vector",
		EmbeddingJSON: memory.MarshalEmbedding([]float32{1, 2}),
		Relevance:     0.5,
	}))

	embedder.Configure(embedder.NewHash())
	checks, err := memory.NewEngine(store).Doctor(context.Background())
	require.NoError(t, err)
	var dimCheck memory.DoctorCheck
	for _, c := range checks {
		if c.Name == "dimensions" {
			dimCheck = c
		}
	}
	require.Error(t, dimCheck.Err)
}

func TestDoctorOrphanLinkedFlow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		ID:          "orphan-1",
		Scope:       runtime.ScopeFlow,
		Type:        "note",
		Summary:     "lost context",
		LinkedFlows: []string{"flow-missing"},
		Relevance:   0.5,
	}))

	embedder.Configure(embedder.NewHash())
	checks, err := memory.NewEngine(store).Doctor(context.Background())
	require.NoError(t, err)
	var orphanCheck memory.DoctorCheck
	for _, c := range checks {
		if c.Name == "orphans" {
			orphanCheck = c
		}
	}
	require.Error(t, orphanCheck.Err)
	require.Contains(t, orphanCheck.Err.Error(), "flow-missing")
}
