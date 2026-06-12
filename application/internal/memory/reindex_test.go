package memory_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestReindexUpdatesEmbeddings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	embedder.Configure(embedder.NewHash())
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope: runtime.ScopeProject, Type: "note", Summary: "alpha beta",
		Relevance: 0.5, EmbeddingJSON: "",
	}))

	n, err := memory.NewEngine(store).Reindex(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, n)

	entries, err := store.ListMemory("", 0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.NotEmpty(t, entries[0].EmbeddingJSON)
}
